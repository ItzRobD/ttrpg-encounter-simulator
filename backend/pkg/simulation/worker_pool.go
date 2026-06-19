package simulation

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/model"
)

// jobTimeout bounds how long a single simulation job may run before its
// context is cancelled, so a pathological payload can't pin a worker forever.
const jobTimeout = 10 * time.Minute

type Job struct {
	SimulationID string
	Payload      MultiSimulationRequest
}

type WorkerPool struct {
	jobs    chan Job
	workers int
	wg      sync.WaitGroup
}

var (
	pool         *WorkerPool
	once         sync.Once
	shutdownOnce sync.Once
)

func InitWorkerPool(workers int) {
	once.Do(func() {
		pool = &WorkerPool{
			jobs:    make(chan Job, 100),
			workers: workers,
		}
		pool.start()
	})
}

func (wp *WorkerPool) start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	slog.Info("worker started", "worker", id)

	for job := range wp.jobs {
		// Process each job in its own func so a panic fails only that
		// simulation, not the worker (and not the process).
		wp.processJob(id, job)
	}
}

func (wp *WorkerPool) processJob(id int, job Job) {
	slog.Info("worker processing simulation", "worker", id, "sim_id", job.SimulationID)

	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	// Recover from any panic in the simulation engine and mark the job failed
	// rather than crashing the worker goroutine.
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("simulation panicked: %v", r)
			slog.Error("simulation panicked", "worker", id, "sim_id", job.SimulationID, "panic", r, "stack", string(debug.Stack()))
			_ = UpdateSimulationStatus(context.Background(), job.SimulationID, model.Status_Failed, &errMsg)
		}
	}()

	// Update status to Running in DB
	if err := UpdateSimulationStatus(ctx, job.SimulationID, model.Status_Running, nil); err != nil {
		slog.Error("failed to update status to running", "worker", id, "sim_id", job.SimulationID, "err", err)
		return
	}

	// Run simulation
	result, err := RunMultiSimulation(ctx, job.Payload)
	if err != nil {
		errMsg := err.Error()
		slog.Error("simulation failed", "worker", id, "sim_id", job.SimulationID, "err", err)
		_ = UpdateSimulationStatus(context.Background(), job.SimulationID, model.Status_Failed, &errMsg)
		return
	}

	// Save results to DB
	if err := UpdateSimulationResult(ctx, job.SimulationID, result); err != nil {
		slog.Error("failed to save result", "worker", id, "sim_id", job.SimulationID, "err", err)
	}

	slog.Info("worker finished simulation", "worker", id, "sim_id", job.SimulationID)
}

func EnqueueSimulation(simID string, payload MultiSimulationRequest) {
	if pool == nil {
		slog.Error("worker pool not initialized")
		return
	}
	pool.jobs <- Job{SimulationID: simID, Payload: payload}
}

// ShutdownWorkerPool stops accepting jobs and waits for in-flight jobs to
// finish. Safe to call more than once (e.g. explicit call + deferred call).
func ShutdownWorkerPool() {
	shutdownOnce.Do(func() {
		if pool != nil {
			close(pool.jobs)
			pool.wg.Wait()
		}
	})
}
