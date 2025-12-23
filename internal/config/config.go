package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	Host          string
}

func NewConfig() *Config {
	cfg := &Config{}
	
	flag.StringVar(&cfg.Host, "a", ":8080", "server host:port to listen on")
	flag.StringVar(&cfg.ServerAddress, "b", "http://localhost:8080", "base URL for shortened links")

	flag.Parse()

	if envHost := os.Getenv("HOST"); envHost != ""{
		cfg.Host = envHost
	}
	if envAddress := os.Getenv("SERVER_ADDRESS"); envAddress != ""{
		cfg.ServerAddress = envAddress
	}
	
	return cfg
}
