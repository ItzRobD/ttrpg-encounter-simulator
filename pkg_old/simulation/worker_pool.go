package simulation

import (
	"context"
	"log"
	"sync"

	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/model"
)

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
	pool *WorkerPool
	once sync.Once
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
	log.Printf("Worker %d started", id)

	for job := range wp.jobs {
		log.Printf("Worker %d processing simulation %s", id, job.SimulationID)

		// Update status to Running
		ctx := context.Background()
		err := UpdateSimulationStatus(ctx, job.SimulationID, model.Status_Running, nil)
		if err != nil {
			log.Printf("Worker %d: failed to update status to running for %s: %v", id, job.SimulationID, err)
			continue
		}

		// Run simulation
		result, err := RunMultiSimulation(ctx, job.Payload)
		if err != nil {
			errMsg := err.Error()
			log.Printf("Worker %d: simulation %s failed: %v", id, job.SimulationID, err)
			// Explicitly log failure for analysis
			log.Printf("Simulation %s failure details: %+v", job.SimulationID, err)
			_ = UpdateSimulationStatus(ctx, job.SimulationID, model.Status_Failed, &errMsg)
			continue
		}

		// Save results
		err = UpdateSimulationResult(ctx, job.SimulationID, result)
		if err != nil {
			log.Printf("Worker %d: failed to save result for %s: %v", id, job.SimulationID, err)
		}

		log.Printf("Worker %d finished simulation %s", id, job.SimulationID)
	}
}

func EnqueueSimulation(simID string, payload MultiSimulationRequest) {
	if pool == nil {
		log.Printf("Worker pool not initialized")
		return
	}
	pool.jobs <- Job{SimulationID: simID, Payload: payload}
}

func ShutdownWorkerPool() {
	if pool != nil {
		close(pool.jobs)
		pool.wg.Wait()
	}
}
