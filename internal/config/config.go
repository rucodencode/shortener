package config

import (
	"errors"
	"flag"
	"os"
)

type RepoType string

const (
	RepoTypePostgres RepoType = "postgres"
	RepoTypeFile     RepoType = "file"
	RepoTypeMemory   RepoType = "memory"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DB              struct {
		DataSourceName string
	}
	RepoType        RepoType
	TokenSigningKey string
}

func NewConfig() (Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DB.DataSourceName, "d", "", "DB data source name")
	flag.StringVar(&cfg.TokenSigningKey, "skey", "12345678123456781234567812345678", "Token signing key")
	flag.Parse()

	envServerAddress, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok && envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	envBaseURL, ok := os.LookupEnv("BASE_URL")
	if ok && envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	envLogLevel, ok := os.LookupEnv("LOG_LEVEL")
	if ok && envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		cfg.FileStoragePath = envFileStoragePath
	}

	envDBDataSourceName, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		cfg.DB.DataSourceName = envDBDataSourceName
	}

	envTokenSigningKey, ok := os.LookupEnv("TOKEN_SIGNING_KEY")
	if ok {
		cfg.TokenSigningKey = envTokenSigningKey
	}

	if len(cfg.TokenSigningKey) < 32 {
		return Config{}, errors.New("token signing key must contain at least 32 bytes")
	}

	if cfg.DB.DataSourceName != "" {
		cfg.RepoType = RepoTypePostgres
	} else if cfg.FileStoragePath != "" {
		cfg.RepoType = RepoTypeFile
	} else {
		cfg.RepoType = RepoTypeMemory
	}

	return cfg, nil
}
