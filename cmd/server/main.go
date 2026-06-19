package main

import (
	"context"
	"log/slog"
	"net/http"
	_ "net/http/pprof" // registers pprof handlers on the default mux (opt-in via ENABLE_PPROF)
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dnd5e-encounter-simulator-backend/internal/api"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/simulation"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// maxRequestBodyBytes caps inbound request bodies (simulation payloads are the
// largest, but still small). Protects against oversized/abusive uploads.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

func main() {
	// pprof is opt-in: only expose the debug server when ENABLE_PPROF is set.
	if os.Getenv("ENABLE_PPROF") == "true" {
		go func() {
			addr := "localhost:6060"
			slog.Info("starting pprof server", "addr", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				slog.Error("pprof server stopped", "err", err)
			}
		}()
	}

	if err := database.InitDb(nil); err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer database.CloseDb()

	simulation.InitWorkerPool(4)
	defer simulation.ShutdownWorkerPool()

	r := gin.Default()
	r.Use(corsMiddleware())
	r.Use(maxBodyMiddleware())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	api.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Run the server in a goroutine so the main goroutine can wait for signals.
	go func() {
		slog.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Block until we receive an interrupt/terminate signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutdown signal received, draining...")

	// Stop accepting new connections and give in-flight requests time to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}

	// Drain the worker pool (waits for in-flight simulations) and close the DB.
	// These also run via defer, but call ShutdownWorkerPool here so simulations
	// finish before the deferred CloseDb runs.
	simulation.ShutdownWorkerPool()
	slog.Info("shutdown complete")
}

// corsMiddleware builds a CORS config from the ALLOWED_ORIGINS env var
// (comma-separated). Falls back to localhost dev origins when unset.
func corsMiddleware() gin.HandlerFunc {
	origins := []string{"http://localhost:4200"}
	if env := os.Getenv("ALLOWED_ORIGINS"); env != "" {
		origins = nil
		for _, o := range strings.Split(env, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = origins
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	cfg.AllowCredentials = true
	return cors.New(cfg)
}

// maxBodyMiddleware caps the size of inbound request bodies.
func maxBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		c.Next()
	}
}
