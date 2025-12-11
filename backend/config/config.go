package config

import (
	"os"
)

// AppConfig holds runtime configuration sourced from environment variables.
type AppConfig struct {
	DBPath    string
	BindAddr  string
	ExportDir string
	Upstream  string
	BranchID  string
}

// Load reads environment variables with sane defaults for desktop sidecar usage.
func Load() AppConfig {
	return AppConfig{
		DBPath:    valueOrDefault("POS_DB_PATH", "offline.db"),
		BindAddr:  valueOrDefault("POS_BIND_ADDR", "0.0.0.0:8080"),
		ExportDir: valueOrDefault("POS_EXPORT_DIR", "exports"),
		Upstream:  valueOrDefault("POS_UPSTREAM_URL", ""),
		BranchID:  valueOrDefault("POS_BRANCH_ID", "local"),
	}
}

func valueOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
