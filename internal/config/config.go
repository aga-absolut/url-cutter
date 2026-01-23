package config

import (
	"flag"
	"math/rand/v2"

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
	flag.StringVar(&cfg.FilePath, "f", "", "storage filename")             // storage.txt
	flag.StringVar(&cfg.DBDSN, "d", "", "name for check connect database") // postgres://postgres:absolute_1@localhost:5432/mydb
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}

func (c Config) Generate() string {
	res := make([]byte, 8)
	for i := range res {
		res[i] = c.Symbols[rand.IntN(len(c.Symbols))]
	}
	return string(res)
}
