package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// loadEnvFile reads .env file and sets environment variables
func loadEnvFile() {
	// Get current executable directory for locating .env
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// Try .env in various locations
	paths := []string{
		".env",
		filepath.Join(exeDir, ".env"),
		filepath.Join(exeDir, "..", ".env"),
		filepath.Join(exeDir, "..", "..", ".env"),
		filepath.Join(exeDir, "..", "..", "..", ".env"),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Parse and set each line
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			// Skip comments and empty lines
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Find the first = sign
			if idx := strings.Index(line, "="); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])

				// Remove surrounding quotes
				value = strings.Trim(value, `"'`)

				// Only set if not already set
				if _, exists := os.LookupEnv(key); !exists {
					os.Setenv(key, value)
				}
			}
		}
		break // Only load first found .env
	}
}

func init() {
	loadEnvFile()
}

type AuthConfig struct {
	JWTSecret      string
	EnableHFAuth   bool
	HFClientID     string
	HFClientSecret string
	HFCallbackURL  string
	EnableLDAP     bool
	LDAPServer     string
	LDAPPort       int
	LDAPBindDN     string
	LDAPBindPass   string
	LDAPBaseDN     string
	LDAPFilter     string
}

type DatabaseConfig struct {
	Type     string // "sqlite" or "postgres"
	Path     string // SQLite database path
	Host     string // PostgreSQL host
	Port     int    // PostgreSQL port
	User     string // PostgreSQL user
	Password string // PostgreSQL password
	Database string // PostgreSQL database name
	SSLMode  string // PostgreSQL SSL mode
}

type StorageConfig struct {
	ModelsPath   string
	DatasetsPath string
	SpacesPath   string
}

type LimitsConfig struct {
	MaxFileSize    int64
	MaxRepoSize    int64
	MaxRequestSize int64
	RequestTimeout time.Duration
}

type RateLimitConfig struct {
	Enabled     bool
	RequestsMin int
	Burst       int
}

type Config struct {
	Port     int
	DataDir  string
	LogLevel string
	Auth     AuthConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Limits   LimitsConfig
	RateLimit RateLimitConfig
}

func Load() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Port, "port", 8080, "Server port")
	flag.StringVar(&cfg.DataDir, "data-dir", "./data", "Data storage directory")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level")

	// Database flags
	flag.StringVar(&cfg.Database.Type, "db-type", "sqlite", "Database type (sqlite or postgres)")
	flag.StringVar(&cfg.Database.Path, "db-path", "", "SQLite database path (default: data-dir/hf-local.db)")
	flag.StringVar(&cfg.Database.Host, "db-host", "localhost", "PostgreSQL host")
	flag.IntVar(&cfg.Database.Port, "db-port", 5432, "PostgreSQL port")
	flag.StringVar(&cfg.Database.User, "db-user", "postgres", "PostgreSQL user")
	flag.StringVar(&cfg.Database.Password, "db-password", "", "PostgreSQL password")
	flag.StringVar(&cfg.Database.Database, "db-name", "hf_local_hub", "PostgreSQL database name")
	flag.StringVar(&cfg.Database.SSLMode, "db-sslmode", "disable", "PostgreSQL SSL mode")

	// Auth flags (removed simple token auth, kept HF OAuth and LDAP)
	flag.BoolVar(&cfg.Auth.EnableHFAuth, "auth-hf", false, "Enable Hugging Face OAuth")
	flag.BoolVar(&cfg.Auth.EnableLDAP, "auth-ldap", false, "Enable LDAP authentication")

	flag.StringVar(&cfg.Auth.HFClientID, "hf-client-id", "", "HF OAuth client ID")
	flag.StringVar(&cfg.Auth.HFClientSecret, "hf-client-secret", "", "HF OAuth client secret")
	flag.StringVar(&cfg.Auth.HFCallbackURL, "hf-callback-url", "http://localhost:8080/auth/hf/callback", "HF OAuth callback URL")

	flag.StringVar(&cfg.Auth.LDAPServer, "ldap-server", "", "LDAP server address")
	flag.IntVar(&cfg.Auth.LDAPPort, "ldap-port", 389, "LDAP server port")
	flag.StringVar(&cfg.Auth.LDAPBindDN, "ldap-bind-dn", "", "LDAP bind DN")
	flag.StringVar(&cfg.Auth.LDAPBindPass, "ldap-bind-pass", "", "LDAP bind password")
	flag.StringVar(&cfg.Auth.LDAPBaseDN, "ldap-base-dn", "", "LDAP base DN")
	flag.StringVar(&cfg.Auth.LDAPFilter, "ldap-filter", "(uid=%s)", "LDAP search filter")

	flag.StringVar(&cfg.Storage.ModelsPath, "storage-models", "", "Models storage path (default: data-dir/storage/models)")
	flag.StringVar(&cfg.Storage.DatasetsPath, "storage-datasets", "", "Datasets storage path (default: data-dir/storage/datasets)")
	flag.StringVar(&cfg.Storage.SpacesPath, "storage-spaces", "", "Spaces storage path (default: data-dir/storage/spaces)")

	flag.Int64Var(&cfg.Limits.MaxFileSize, "max-file-size", 10*1024*1024*1024, "Max file size in bytes")
	flag.Int64Var(&cfg.Limits.MaxRepoSize, "max-repo-size", 100*1024*1024*1024, "Max repo size in bytes")
	flag.Int64Var(&cfg.Limits.MaxRequestSize, "max-request-size", 10*1024*1024*1024, "Max request size in bytes")
	flag.DurationVar(&cfg.Limits.RequestTimeout, "request-timeout", 30*time.Minute, "Request timeout")

	flag.BoolVar(&cfg.RateLimit.Enabled, "rate-limit", false, "Enable rate limiting")
	flag.IntVar(&cfg.RateLimit.RequestsMin, "rate-limit-rpm", 60, "Rate limit requests per minute")
	flag.IntVar(&cfg.RateLimit.Burst, "rate-limit-burst", 10, "Rate limit burst size")

	flag.Parse()

	cfg.loadFromEnv()
	cfg.setDefaults()
	cfg.setJWTSecret()

	return cfg
}

func (cfg *Config) loadFromEnv() {
	if port := os.Getenv("HF_LOCAL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}
	if dir := os.Getenv("HF_LOCAL_DATA_DIR"); dir != "" {
		cfg.DataDir = dir
	}
	if level := os.Getenv("HF_LOCAL_LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}
	if secret := os.Getenv("HF_LOCAL_JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
	}

	// Database configuration from env
	if dbType := os.Getenv("HF_LOCAL_DB_TYPE"); dbType != "" {
		cfg.Database.Type = dbType
	}
	if dbPath := os.Getenv("HF_LOCAL_DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if host := os.Getenv("HF_LOCAL_DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("HF_LOCAL_DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("HF_LOCAL_DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if pass := os.Getenv("HF_LOCAL_DB_PASSWORD"); pass != "" {
		cfg.Database.Password = pass
	}
	if db := os.Getenv("HF_LOCAL_DB_NAME"); db != "" {
		cfg.Database.Database = db
	}
	if sslMode := os.Getenv("HF_LOCAL_DB_SSLMODE"); sslMode != "" {
		cfg.Database.SSLMode = sslMode
	}

	if os.Getenv("HF_LOCAL_AUTH_HF") == "true" {
		cfg.Auth.EnableHFAuth = true
	}
	if id := os.Getenv("HF_LOCAL_HF_CLIENT_ID"); id != "" {
		cfg.Auth.HFClientID = id
	}
	if secret := os.Getenv("HF_LOCAL_HF_CLIENT_SECRET"); secret != "" {
		cfg.Auth.HFClientSecret = secret
	}
	if url := os.Getenv("HF_LOCAL_HF_CALLBACK_URL"); url != "" {
		cfg.Auth.HFCallbackURL = url
	}
	if os.Getenv("HF_LOCAL_AUTH_LDAP") == "true" {
		cfg.Auth.EnableLDAP = true
	}
	if server := os.Getenv("HF_LOCAL_LDAP_SERVER"); server != "" {
		cfg.Auth.LDAPServer = server
	}
	if port := os.Getenv("HF_LOCAL_LDAP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Auth.LDAPPort = p
		}
	}
	if dn := os.Getenv("HF_LOCAL_LDAP_BIND_DN"); dn != "" {
		cfg.Auth.LDAPBindDN = dn
	}
	if pass := os.Getenv("HF_LOCAL_LDAP_BIND_PASS"); pass != "" {
		cfg.Auth.LDAPBindPass = pass
	}
	if base := os.Getenv("HF_LOCAL_LDAP_BASE_DN"); base != "" {
		cfg.Auth.LDAPBaseDN = base
	}
	if filter := os.Getenv("HF_LOCAL_LDAP_FILTER"); filter != "" {
		cfg.Auth.LDAPFilter = filter
	}
	if path := os.Getenv("HF_LOCAL_STORAGE_MODELS"); path != "" {
		cfg.Storage.ModelsPath = path
	}
	if path := os.Getenv("HF_LOCAL_STORAGE_DATASETS"); path != "" {
		cfg.Storage.DatasetsPath = path
	}
	if path := os.Getenv("HF_LOCAL_STORAGE_SPACES"); path != "" {
		cfg.Storage.SpacesPath = path
	}
	if size := os.Getenv("HF_LOCAL_MAX_FILE_SIZE"); size != "" {
		if s, err := strconv.ParseInt(size, 10, 64); err == nil {
			cfg.Limits.MaxFileSize = s
		}
	}
	if size := os.Getenv("HF_LOCAL_MAX_REPO_SIZE"); size != "" {
		if s, err := strconv.ParseInt(size, 10, 64); err == nil {
			cfg.Limits.MaxRepoSize = s
		}
	}
	if size := os.Getenv("HF_LOCAL_MAX_REQUEST_SIZE"); size != "" {
		if s, err := strconv.ParseInt(size, 10, 64); err == nil {
			cfg.Limits.MaxRequestSize = s
		}
	}
	if timeout := os.Getenv("HF_LOCAL_REQUEST_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			cfg.Limits.RequestTimeout = t
		}
	}
	if os.Getenv("HF_LOCAL_RATE_LIMIT") == "true" {
		cfg.RateLimit.Enabled = true
	}
	if rpm := os.Getenv("HF_LOCAL_RATE_LIMIT_RPM"); rpm != "" {
		if r, err := strconv.Atoi(rpm); err == nil {
			cfg.RateLimit.RequestsMin = r
		}
	}
	if burst := os.Getenv("HF_LOCAL_RATE_LIMIT_BURST"); burst != "" {
		if b, err := strconv.Atoi(burst); err == nil {
			cfg.RateLimit.Burst = b
		}
	}
}

func (cfg *Config) setDefaults() {
	if cfg.Database.Path == "" && cfg.Database.Type == "sqlite" {
		cfg.Database.Path = cfg.DataDir + "/hf-local.db"
	}
	if cfg.Storage.ModelsPath == "" {
		cfg.Storage.ModelsPath = cfg.DataDir + "/storage/models"
	}
	if cfg.Storage.DatasetsPath == "" {
		cfg.Storage.DatasetsPath = cfg.DataDir + "/storage/datasets"
	}
	if cfg.Storage.SpacesPath == "" {
		cfg.Storage.SpacesPath = cfg.DataDir + "/storage/spaces"
	}
}

// generateRandomSecret generates a cryptographically secure random hex string
func generateRandomSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (cfg *Config) setJWTSecret() {
	if cfg.Auth.JWTSecret == "" {
		// Generate a random secret for development/testing
		// In production, always set HF_LOCAL_JWT_SECRET environment variable
		secret, err := generateRandomSecret()
		if err != nil {
			log.Printf("Warning: Failed to generate random JWT secret: %v", err)
			cfg.Auth.JWTSecret = "change-me-in-production"
		} else {
			cfg.Auth.JWTSecret = secret
			log.Printf("Generated random JWT secret (set HF_LOCAL_JWT_SECRET for persistence)")
		}
	}
}
