package env

import (
	"log/slog"
	"os"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	slog.Warn("Env variable not found, resolving to default", "key", key, "default", defaultValue)
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	slog.Warn("Env variable not found, resolving to default", "key", key, "default", defaultValue)
	return defaultValue
}

func GetEnvBool(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	slog.Warn("Env variable not found, resolving to default", "key", key, "default", defaultValue)
	return defaultValue
}
