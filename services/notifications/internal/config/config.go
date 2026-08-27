package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

const (
	defaultHost       = "0.0.0.0"
	defaultPort       = "8081"
	defaultBackendURL = "https://localhost:8080/api/v1"
	defaultDBPath     = "db/data/notifications.db"
	defaultCORSOrigin = "http://localhost:3001"

	// SSE streams are long-lived responses — a non-zero write timeout would
	// kill every active stream once it elapsed.
	defaultWriteTimeout = 0
)

type Config struct {
	Host           string
	Port           string
	Backend        string
	DBPath         string
	AllowedOrigins string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func Load() (*Config, error) {
	host := env("NOTIFICATIONS_HOST", defaultHost)
	port := env("NOTIFICATIONS_PORT", defaultPort)
	backend := env("NOTIFICATIONS_BACKEND_URL", defaultBackendURL)
	dbPath := env("NOTIFICATIONS_DB_PATH", defaultDBPath)
	allowedOrigins := env("NOTIFICATIONS_ALLOWED_ORIGINS", defaultCORSOrigin)

	readTimeout, err := parseInt("NOTIFICATIONS_READ_TIMEOUT", 10)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := parseInt("NOTIFICATIONS_WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return nil, err
	}
	idleTimeout, err := parseInt("NOTIFICATIONS_IDLE_TIMEOUT", 30)
	if err != nil {
		return nil, err
	}

	if host == "" {
		return nil, errors.New("missing NOTIFICATIONS_HOST")
	}
	if port == "" {
		return nil, errors.New("missing NOTIFICATIONS_PORT")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return nil, errors.New("NOTIFICATIONS_PORT must be an integer")
	}

	return &Config{
		Host:           host,
		Port:           port,
		Backend:        backend,
		DBPath:         dbPath,
		AllowedOrigins: allowedOrigins,
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
		IdleTimeout:    time.Duration(idleTimeout) * time.Second,
	}, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return n, nil
}
