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

func NewConfig() *Config {
	cfg := &Config{}
	cfg.Symbols = []byte("QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm")
	
	flag.StringVar(&cfg.Host, "a", ":8080", "server host:port")
	flag.StringVar(&cfg.ServerAddress, "b", "http://localhost:8080/", "base URL")
	flag.StringVar(&cfg.FilePath, "f", "storage.txt", "storage filename")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}
