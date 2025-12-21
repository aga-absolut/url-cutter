package config


type Config struct{
	ServerAddress string
	Host string
}

func NewConfig() *Config {
	return &Config{}
}