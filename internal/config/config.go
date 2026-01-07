package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress string `env:"BASE_URL"`
	Host          string `env:"SERVER_ADDRESS"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Symbols       []byte
}

func NewConfig() Config {
	var cfg Config
	cfg.Symbols = []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	if cfg.Host != "" && cfg.ServerAddress != "" {
		return cfg
	}
	if cfg.FilePath == "" {
		flag.StringVar(&cfg.FilePath, "f", "storage.txt", "path to file")
	}

	flag.StringVar(&cfg.Host, "a", ":8080", "server host:port to listen on")
	flag.StringVar(&cfg.ServerAddress, "b", "http://localhost:8080/", "base URL for shortened links")
	flag.StringVar(&cfg.FilePath, "f", "storage.txt", "storage filename")

	return cfg
}
