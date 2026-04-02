package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"
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
	ServerAddress string `json:"server_address" env:"SERVER_ADDRESS"`
	Host          string `json:"base_url" env:"BASE_URL"`
	FilePath      string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DBDSN         string `json:"database_dsn" env:"DATABASE_DSN"`
	ConfigFile    string `json:"-" env:"CONFIG"`
	EnableHTTPS   bool   `json:"enable_https" env:"ENABLE_HTTPS"`
}

// NewConfig создает новый Config со значениями из флагов или переменных окружения
func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ConfigFile, "c", "", "config file name")           // config.txt
	flag.StringVar(&cfg.ServerAddress, "a", "", "server host:port")        // localhost:8080
	flag.StringVar(&cfg.Host, "b", "", "base URL")                         // http://localhost:8080
	flag.StringVar(&cfg.FilePath, "f", "", "storage filename")             // storage.txt
	flag.StringVar(&cfg.DBDSN, "d", "", "name for check connect database") // postgres://postgres:absolute_1@localhost:5432/mydb
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "https")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	if cfg.ConfigFile != "" {
		if err := ParseConfigFromFile(cfg.ConfigFile, cfg); err != nil {
			panic(err)
		}
	}

	return cfg
}

// ParseConfigFromFile парсит данные для конфига из файла
func ParseConfigFromFile(name string, cfg *Config) error {
	configFile := &Config{}

	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &configFile); err != nil {
		return err
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = configFile.ServerAddress
	}
	if cfg.Host == "" {
		cfg.Host = configFile.Host
	}
	if cfg.FilePath == "" {
		cfg.FilePath = configFile.FilePath
	}
	if cfg.DBDSN == "" {
		cfg.DBDSN = configFile.DBDSN
	}
	if !cfg.EnableHTTPS {
		cfg.EnableHTTPS = configFile.EnableHTTPS
	}

	return nil
}
