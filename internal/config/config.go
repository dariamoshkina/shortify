package config

import (
	"crypto/aes"
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/dariamoshkina/shortify/internal"
	"github.com/samber/lo"
)

const (
	DefaultServerAddress = "localhost:8080"
	DefaultBaseURL       = "http://localhost:8080"
	DefaultFilename      = "urls.json"
)

type Config struct {
	Addr          string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileStorage   string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
	EncryptionKey []byte
}

func Init() (*Config, error) {
	var (
		config                                  Config
		addr, baseURL, fileStorage, databaseDSN *string
	)

	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	addr = flag.String("a", DefaultServerAddress, "host URL")
	baseURL = flag.String("b", DefaultBaseURL, "base URL")
	fileStorage = flag.String("f", DefaultFilename, "file storage path")
	databaseDSN = flag.String("d", "", "database DSN")
	flag.Parse()

	config.Addr, _ = lo.Coalesce(config.Addr, *addr)
	config.BaseURL, _ = lo.Coalesce(config.BaseURL, *baseURL)
	config.FileStorage, _ = lo.Coalesce(config.FileStorage, *fileStorage)
	config.DatabaseDSN, _ = lo.Coalesce(config.DatabaseDSN, *databaseDSN)

	encryptionKey, err := internal.RandomBytes(2 * aes.BlockSize)
	if err != nil {
		return nil, err
	}

	config.EncryptionKey = encryptionKey

	return &config, nil
}
