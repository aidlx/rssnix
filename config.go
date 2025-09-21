package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
)

const (
	configEnvVar         = "RSSNIX_CONFIG_HOME"
	defaultConfigDir     = ".config/rssnix"
	configFileName       = "config.ini"
	defaultViewer        = "vim"
	defaultFeedDirectory = "~/rssnix"
)

type Configuration struct {
	FeedDirectory string
	Viewer        string
	Feeds         []Feed
}

var Config Configuration

func LoadConfig() error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	cfgDir, err := resolveConfigDir(homePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return fmt.Errorf("ensure config directory %q: %w", cfgDir, err)
	}

	cfgPath := filepath.Join(cfgDir, configFileName)
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		log.Warn("Config file does not exist, creating...")
		if err := createDefaultConfig(cfgPath); err != nil {
			return fmt.Errorf("create default config: %w", err)
		}
		log.Infof("Config file created at %s", cfgPath)
	} else if err != nil {
		return fmt.Errorf("stat config file: %w", err)
	}

	cfg, err := ini.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	Config = Configuration{}
	settings := cfg.Section("settings")

	feedDirValue := strings.TrimSpace(settings.Key("feed_directory").String())
	if feedDirValue == "" {
		feedDirValue = defaultFeedDirectory
	}
	Config.FeedDirectory = expandPath(feedDirValue, homePath)
	if err := os.MkdirAll(Config.FeedDirectory, 0o755); err != nil {
		return fmt.Errorf("ensure feed directory %q: %w", Config.FeedDirectory, err)
	}

	viewer := strings.TrimSpace(settings.Key("viewer").String())
	if viewer == "" {
		viewer = defaultViewer
	}
	Config.Viewer = viewer

	feedsSection := cfg.Section("feeds")
	for _, key := range feedsSection.Keys() {
		name := strings.TrimSpace(key.Name())
		if name == "" {
			continue
		}

		url := strings.TrimSpace(key.String())
		if url == "" {
			log.WithField("feed", name).Warn("Feed has empty URL; skipping")
			continue
		}

		Config.Feeds = append(Config.Feeds, Feed{Name: name, URL: url})
	}

	if len(Config.Feeds) == 0 {
		log.Warn("No feeds configured; use `rssnix add` or `rssnix import` to add feeds")
	}

	return nil
}

func resolveConfigDir(home string) (string, error) {
	override := strings.TrimSpace(os.Getenv(configEnvVar))
	if override == "" {
		return filepath.Join(home, defaultConfigDir), nil
	}

	override = expandPath(override, home)
	if !filepath.IsAbs(override) {
		override = filepath.Join(home, override)
	}
	return override, nil
}

func expandPath(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func createDefaultConfig(path string) error {
	cfg := ini.Empty()
	settings := cfg.Section("settings")
	settings.Key("viewer").SetValue(defaultViewer)
	settings.Key("feed_directory").SetValue(defaultFeedDirectory)
	cfg.Section("feeds")
	return cfg.SaveTo(path)
}

func configFilePath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	cfgDir, err := resolveConfigDir(homePath)
	if err != nil {
		return "", err
	}

	return filepath.Join(cfgDir, configFileName), nil
}

func (c *Configuration) FeedByName(name string) (Feed, bool) {
	for _, feed := range c.Feeds {
		if feed.Name == name {
			return feed, true
		}
	}
	return Feed{}, false
}
