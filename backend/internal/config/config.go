package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	AllowedOrigins  []string
	LocationRPS     float64
	LocationBurst   int
	DefaultPageSize int
	MaxPageSize     int
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AllowedOrigins:  []string{getenv("CORS_ALLOWED_ORIGIN", "*")},
		LocationRPS:     getenvFloat("LOCATION_RPS", 30),
		LocationBurst:   getenvInt("LOCATION_BURST", 60),
		DefaultPageSize: getenvInt("DEFAULT_PAGE_SIZE", 50),
		MaxPageSize:     getenvInt("MAX_PAGE_SIZE", 500),
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	return c, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getenvInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func getenvFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return d
}
