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
	DefaultFilename      = "urls.json"
)

type Config struct {
	Addr        string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
}

func Parse() *Config {
	var (
		config                     Config
		addr, baseURL, fileStorage *string
	)

	if err := env.Parse(&config); err != nil {
		fmt.Println(err)
	}

	addr = flag.String("a", DefaultServerAddress, "host URL")
	baseURL = flag.String("b", DefaultBaseURL, "base URL")
	fileStorage = flag.String("f", DefaultFilename, "file storage path")
	flag.Parse()

	config.Addr, _ = lo.Coalesce(config.Addr, *addr)
	config.BaseURL, _ = lo.Coalesce(config.BaseURL, *baseURL)
	config.FileStorage, _ = lo.Coalesce(config.FileStorage, *fileStorage)

	return &config
}
