package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Akicou/hf-local-hub/server/api"
	"github.com/Akicou/hf-local-hub/server/config"
	"github.com/Akicou/hf-local-hub/server/db"
	"github.com/Akicou/hf-local-hub/server/middleware"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	if cfg.LogLevel == "debug" {
		logger, err = zap.NewDevelopment()
		if err != nil {
			log.Fatalf("Failed to initialize debug logger: %v", err)
		}
	}

	logger.Info("Starting hf-local-hub server",
		zap.Int("port", cfg.Port),
		zap.String("data_dir", cfg.DataDir),
	)

	database, err := db.InitDB(cfg.DataDir + "/hf-local.db")
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func() {
		if err := db.CloseDB(database); err != nil {
			logger.Error("Failed to close database", zap.Error(err))
		}
	}()

	server := api.New(cfg, database, logger)
	router := server.SetupRouter()

	var rateLimiter *middleware.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = middleware.NewRateLimiter(cfg.RateLimit.RequestsMin, cfg.RateLimit.Burst)
		router.Use(rateLimiter.Middleware())
		logger.Info("Rate limiting enabled",
			zap.Int("requests_per_minute", cfg.RateLimit.RequestsMin),
			zap.Int("burst", cfg.RateLimit.Burst),
		)
	}

	if cfg.Limits.MaxFileSize > 0 || cfg.Limits.MaxRequestSize > 0 {
		limits := middleware.LimitsConfig{
			MaxFileSize:    cfg.Limits.MaxFileSize,
			MaxRequestSize: cfg.Limits.MaxRequestSize,
		}
		router.Use(middleware.NewSizeLimits(limits))
		logger.Info("Size limits enabled",
			zap.Int64("max_file_size", cfg.Limits.MaxFileSize),
			zap.Int64("max_request_size", cfg.Limits.MaxRequestSize),
		)
	}

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.Limits.RequestTimeout,
		WriteTimeout: cfg.Limits.RequestTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Stop rate limiter cleanup goroutine
	if rateLimiter != nil {
		rateLimiter.Stop()
	}

	logger.Info("Server stopped")
}
