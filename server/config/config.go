package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type AuthConfig struct {
	JWTSecret      string
	EnableTokenAuth bool
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

type StorageConfig struct {
	ModelsPath  string
	DatasetsPath string
	SpacesPath  string
}

type LimitsConfig struct {
	MaxFileSize      int64
	MaxRepoSize      int64
	MaxRequestSize   int64
	RequestTimeout   time.Duration
}

type RateLimitConfig struct {
	Enabled     bool
	RequestsMin int
	Burst       int
}

type Config struct {
	Port     int
	DataDir  string
	Token    string
	LogLevel string
	Auth     AuthConfig
	Storage  StorageConfig
	Limits   LimitsConfig
	RateLimit RateLimitConfig
}

func Load() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Port, "port", 8080, "Server port")
	flag.StringVar(&cfg.DataDir, "data-dir", "./data", "Data storage directory")
	flag.StringVar(&cfg.Token, "token", "", "Optional authentication token")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level")

	flag.BoolVar(&cfg.Auth.EnableTokenAuth, "auth-token", false, "Enable token authentication")
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
	if token := os.Getenv("HF_LOCAL_TOKEN"); token != "" {
		cfg.Token = token
	}
	if level := os.Getenv("HF_LOCAL_LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}
	if secret := os.Getenv("HF_LOCAL_JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
	}
	if os.Getenv("HF_LOCAL_AUTH_TOKEN") == "true" {
		cfg.Auth.EnableTokenAuth = true
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

func (cfg *Config) setJWTSecret() {
	if cfg.Auth.JWTSecret == "" {
		if cfg.Token != "" {
			cfg.Auth.JWTSecret = cfg.Token
		} else {
			cfg.Auth.JWTSecret = "change-me-in-production"
		}
	}
}
