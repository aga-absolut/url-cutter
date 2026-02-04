package config

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/caarlos0/env/v11"
)

var (
	SecretKey    = []byte("my_secret_key")
	TokenExpTime = time.Hour * 3
	SizeWorkers  = 1
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	Host          string `env:"BASE_URL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	DBDSN         string `env:"DATABASE_DSN"`
}

func NewConfig() *Config {
	cfg := &Config{}
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

func Generate(originalURL string) string {
	salt := rand.Intn(9999)
	data := sha256.Sum224([]byte(fmt.Sprintf("%s%d", originalURL, salt)))
	return hex.EncodeToString(data[:4])
}
