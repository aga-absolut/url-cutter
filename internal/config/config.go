package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	ServerAddress string `env:"BASE_URL"`
	Host          string `env:"SERVER_ADDRESS"`
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Host, "a", ":8080", "server host:port to listen on")
	flag.StringVar(&cfg.ServerAddress, "b", "http://localhost:8080", "base URL for shortened links")

	flag.Parse()

	if envHost := os.Getenv("SERVER_ADDRESS"); envHost != "" {
		cfg.Host = envHost
	}
	if envAddress := os.Getenv("BASE_URL"); envAddress != "" {	
		envAddress = strings.Replace(envAddress, "http//" , "http://" , 1)

		if !strings.HasPrefix(envAddress, "http://"){
			cfg.ServerAddress = "http://" + envAddress
		}
		
		if !strings.HasSuffix(envAddress, "/"){
			cfg.ServerAddress = envAddress + "/"
		}
	}

	return cfg
}
