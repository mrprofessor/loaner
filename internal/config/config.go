package config

import "os"

type Config struct {
	GRPCPort    string
	HTTPPort    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://loaner:loaner@localhost:5432/loaner?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
