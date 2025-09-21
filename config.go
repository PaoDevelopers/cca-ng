package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Database struct {
		URL string `toml:"url"`
	} `toml:"database"`

	Server struct {
		Address     string `toml:"address"`
		StaticDir   string `toml:"static_dir"`
		SessionName string `toml:"session_name"`
	} `toml:"server"`

	Schema struct {
		Version int64 `toml:"version"`
	} `toml:"schema"`
}

func loadConfig() (*Config, error) {
	configPath := flag.String("config", "config.toml", "path to configuration file")
	flag.Parse()

	var config Config

	// Set defaults
	config.Server.Address = ":8080"
	config.Server.StaticDir = "./frontend/dist"
	config.Server.SessionName = "cca_session"
	config.Schema.Version = 1

	if _, err := os.Stat(*configPath); err == nil {
		if _, err := toml.DecodeFile(*configPath, &config); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	return &config, nil
}
