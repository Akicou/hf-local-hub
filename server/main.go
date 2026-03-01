package main

import (
	"github.com/lyani/hf-local-hub/server/api"
	"github.com/lyani/hf-local-hub/server/config"
	"github.com/lyani/hf-local-hub/server/db"
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

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("Server listening", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
