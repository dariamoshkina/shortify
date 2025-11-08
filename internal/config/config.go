package config

import "flag"

type Config struct {
	Addr    string
	BaseURL string
}

func Parse() *Config {
	addr := flag.String("a", "localhost:8080", "host URL")
	baseURL := flag.String("b", "http://localhost:8080", "base URL")

	flag.Parse()

	return &Config{
		Addr:    *addr,
		BaseURL: *baseURL,
	}
}
