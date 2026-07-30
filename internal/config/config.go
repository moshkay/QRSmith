// Package config loads runtime configuration from environment variables.
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for the QRForge server.
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxContentBytes int   // max length of the content encoded in a QR code
	MaxLogoBytes    int64 // max upload size for a logo image

	// MongoURI enables persistent link storage when set. When empty, the
	// shortener falls back to an in-memory store (links lost on restart).
	MongoURI     string
	MongoDB      string        // database name for short links
	MongoTimeout time.Duration // per-operation timeout for Mongo calls
}

// Load reads configuration from the environment, applying sane defaults. It
// first loads a local .env file (if present) so `go run` picks up settings the
// same way Docker Compose does. Real environment variables always take
// precedence over .env values.
func Load() Config {
	loadDotEnv(".env")

	return Config{
		Port:            getEnv("PORT", "8080"),
		ReadTimeout:     getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
		MaxContentBytes: getIntEnv("MAX_CONTENT_BYTES", 2048),
		MaxLogoBytes:    int64(getIntEnv("MAX_LOGO_BYTES", 2*1024*1024)), // 2 MiB
		MongoURI:        getEnv("MONGO_URI", ""),
		MongoDB:         getEnv("MONGO_DB", "qrunex"),
		MongoTimeout:    getDurationEnv("MONGO_TIMEOUT", 5*time.Second),
	}
}

// loadDotEnv reads simple KEY=VALUE lines from an .env file and sets any keys
// not already present in the environment. Missing file is not an error. Values
// may be optionally wrapped in single or double quotes. Real env vars win.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // no .env is fine
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}
