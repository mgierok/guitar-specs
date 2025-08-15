package app

import (
	"fmt"
	"os"
)

type Config struct {
	Host string
	Port string
	Env  string
}

func LoadConfig() Config {
	return Config{
		Host: getenv("HOST", "0.0.0.0"),
		Port: getenv("PORT", "8080"),
		Env:  getenv("ENV", "development"),
	}
}

func (c Config) Addr() string { return fmt.Sprintf("%s:%s", c.Host, c.Port) }

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
