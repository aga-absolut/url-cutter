package config

import "flag"

type Config struct {
	ServerAddress string
	Host          string
}

func NewConfig() *Config {
	flA, flB := FlagA()
	return &Config{
		Host:          flA,
		ServerAddress: flB,
	}
}

func FlagA() (string, string) {
	FlagA := flag.String("a", ":8080", "host:port")
	FlagB := flag.String("b", "http://localhost"+*FlagA+"/", "serverAddress")

	flag.Parse()
	
	return *FlagA, *FlagB
}
