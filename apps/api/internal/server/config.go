package server

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	CodexCLIPath        string
	CodexModel          string
	CodexWorkingDir     string
	CodexTimeoutSeconds int
	PythonBin           string
}

func LoadConfig() Config {
	return Config{
		Port:                env("PORT", "8000"),
		DatabaseURL:         env("DATABASE_URL", "sqlite:///./algo_sensei.db"),
		CodexCLIPath:        env("CODEX_CLI_PATH", "codex"),
		CodexModel:          os.Getenv("CODEX_MODEL"),
		CodexWorkingDir:     env("CODEX_WORKDIR", "."),
		CodexTimeoutSeconds: envInt("CODEX_CLI_TIMEOUT_SECONDS", 180),
		PythonBin:           env("PYTHON_BIN", "python3"),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
