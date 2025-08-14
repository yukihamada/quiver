package config

type Config struct {
	Port        string
	StoragePath string
}

func DefaultConfig() *Config {
	return &Config{
		Port:        "8081",
		StoragePath: "./data",
	}
}
