package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
)

// Секретный ключ
var SecretKey = []byte("my_secret_key")

// Время жизни токена
var TokenExpTime = time.Hour * 3

// Количество работающих
var SizeWorkers = 1

// Config структура
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	Host          string `env:"BASE_URL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	DBDSN         string `env:"DATABASE_DSN"`
}

// NewConfig создает новый Config со значениями из флагов или переменных окружения
func NewConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server host:port")
	flag.StringVar(&cfg.Host, "b", "http://localhost:8080", "base URL")
	flag.StringVar(&cfg.FilePath, "f", "", "storage filename")                                                               // storage.txt
	flag.StringVar(&cfg.DBDSN, "d", "postgres://postgres:absolute_1@localhost:5432/mydb", "name for check connect database") // psql -U postgres -d mydb -W
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}
