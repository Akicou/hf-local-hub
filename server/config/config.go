package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Port     int
	DataDir  string
	Token    string
	LogLevel string
}

func Load() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Port, "port", 8080, "Server port")
	flag.StringVar(&cfg.DataDir, "data-dir", "./data", "Data storage directory")
	flag.StringVar(&cfg.Token, "token", "", "Optional authentication token")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	if port := os.Getenv("HF_LOCAL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if dir := os.Getenv("HF_LOCAL_DATA_DIR"); dir != "" {
		cfg.DataDir = dir
	}

	if token := os.Getenv("HF_LOCAL_TOKEN"); token != "" {
		cfg.Token = token
	}

	if level := os.Getenv("HF_LOCAL_LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}

	return cfg
}
