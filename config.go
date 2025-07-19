package main

import (
	"errors"
	"os"

	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
)

var defaultConfigContent = `[settings]
feed_directory = ~/rssnix
viewer = vim
max_concurrent_fetch = 5

[feeds]`

type Config struct {
	FeedDirectory      string
	Viewer             string
	MaxConcurrentFetch int
	Feeds              []Feed
}

func LoadConfig() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Error("Failed to get home path")
		os.Exit(1)
	}

	if _, err := os.Stat(homePath + "/.config/rssnix/config.ini"); errors.Is(err, os.ErrNotExist) {
		log.Warn("Config file does not exist, creating...")

		os.MkdirAll(homePath+"/.config/rssnix", 0744)
		file, err := os.Create(homePath + "/.config/rssnix/config.ini")
		if err != nil {
			log.Error("Failed to create a config file")
			os.Exit(1)
		}
		defer file.Close()

		_, err = file.WriteString(defaultConfigContent)
		if err != nil {
			log.Error("Failed to create a config file")
			os.Exit(1)
		}

		log.Info("Config file created at " + homePath + "/.config/rssnix/config.ini")
	}

	cfg, err := ini.Load(homePath + "/.config/rssnix/config.ini")
	log.Info(cfg)
}
