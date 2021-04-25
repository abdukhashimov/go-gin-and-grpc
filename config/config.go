package config

import (
	"os"

	"github.com/spf13/cast"
)

type Config struct {
	LogLevel string
	HttpPort string

	TodoServiceHost string
	TodoServicePort int
}

func Load() Config {
	config := Config{}

	config.LogLevel = cast.ToString(getOrReturnDefault("LOG_LEVEL", "debug"))
	config.HttpPort = cast.ToString(getOrReturnDefault("HTTP_PORT", ":8080"))

	config.TodoServiceHost = cast.ToString(getOrReturnDefault("TODO_SERICE_HOST", "localhost"))
	config.TodoServicePort = cast.ToInt(getOrReturnDefault("TODO_SERVICE_PORT", 8001))

	return config
}

func getOrReturnDefault(key string, defaultValue interface{}) interface{} {
	_, exists := os.LookupEnv(key)
	if exists {
		return os.Getenv(key)
	}

	return defaultValue
}
