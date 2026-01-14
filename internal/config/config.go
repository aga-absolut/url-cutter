package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	Host          string `env:"BASE_URL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	DBDSN         string `env:"DATABASE_DSN"`
	Symbols       []byte
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Symbols = []byte("QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server host:port")
	flag.StringVar(&cfg.Host, "b", "http://localhost:8080", "base URL")
	flag.StringVar(&cfg.FilePath, "f", "storage.txt", "storage filename")
	flag.StringVar(&cfg.DBDSN, "d", "postgres://postgres:12345@localhost:5432/SQLForShortenURL", "name for check connect database")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}
