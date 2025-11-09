package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/samber/lo"
)

const (
	DefaultServerAddress = "localhost:8080"
	DefaultBaseURL       = "http://localhost:8080"
)

type Config struct {
	Addr    string
	BaseURL string
}

func Parse() *Config {
	var (
		config        Config
		addr, baseURL *string
	)

	if err := env.Parse(&config); err != nil {
		fmt.Println(err)
	}

	addr = flag.String("a", DefaultServerAddress, "host URL")
	baseURL = flag.String("b", DefaultBaseURL, "base URL")
	flag.Parse()

	config.Addr, _ = lo.Coalesce(config.Addr, *addr)
	config.BaseURL, _ = lo.Coalesce(config.BaseURL, *baseURL)

	return &config
}
