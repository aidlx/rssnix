package main

import (
	"errors"
	"os"
	"strings"

	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
)

var defaultConfigContent = `[settings]
viewer = vim
feed_directory = ~/rssnix

[feeds]`

type Configuration struct {
	FeedDirectory string
	Viewer        string
	Feeds         []Feed
}

var Config Configuration

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

	Config = Configuration{}
	Config.FeedDirectory = cfg.Section("settings").Key("feed_directory").String()
	if strings.HasPrefix(Config.FeedDirectory, "~") {
		Config.FeedDirectory = homePath + Config.FeedDirectory[1:]
	}
	os.MkdirAll(Config.FeedDirectory, 0744)
	Config.Viewer = cfg.Section("settings").Key("viewer").String()
	for _, key := range cfg.Section("feeds").Keys() {
		Config.Feeds = append(Config.Feeds, Feed{key.String(), cfg.Section("feeds").Key(key.String()).String()})
	}
}
