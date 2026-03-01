package main

import (
	"github.com/lyani/hf-local-hub/server/api"
	"github.com/lyani/hf-local-hub/server/config"
	"github.com/lyani/hf-local-hub/server/db"
	"github.com/lyani/hf-local-hub/server/middleware"
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
	defer logger.Sync()

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

	server := api.New(cfg, database, logger)
	router := server.SetupRouter()

	if cfg.RateLimit.Enabled {
		rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsMin, cfg.RateLimit.Burst)
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

	logger.Info("Server listening", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
