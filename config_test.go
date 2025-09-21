package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-ini/ini"
)

func TestLoadConfigCreatesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(configEnvVar, "")

	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	cfgPath, err := configFilePath()
	if err != nil {
		t.Fatalf("configFilePath returned error: %v", err)
	}

	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config file not created at %s: %v", cfgPath, err)
	}

	expectedFeedDir := filepath.Join(home, "rssnix")
	if Config.FeedDirectory != expectedFeedDir {
		t.Fatalf("expected feed directory %s, got %s", expectedFeedDir, Config.FeedDirectory)
	}

	if _, err := os.Stat(Config.FeedDirectory); err != nil {
		t.Fatalf("feed directory not created: %v", err)
	}

	if Config.Viewer != defaultViewer {
		t.Fatalf("expected viewer %s, got %s", defaultViewer, Config.Viewer)
	}

	if len(Config.Feeds) != 0 {
		t.Fatalf("expected no feeds configured by default, got %d", len(Config.Feeds))
	}
}

func TestLoadConfigRespectsEnvOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	overrideDir := filepath.Join(home, "custom-config")
	t.Setenv(configEnvVar, overrideDir)

	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	cfgPath, err := configFilePath()
	if err != nil {
		t.Fatalf("configFilePath returned error: %v", err)
	}

	expectedPath := filepath.Join(overrideDir, configFileName)
	if cfgPath != expectedPath {
		t.Fatalf("expected config file path %s, got %s", expectedPath, cfgPath)
	}

	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("config file not created at %s: %v", expectedPath, err)
	}
}

func TestAddFeedPersistsConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(configEnvVar, "")

	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	const name = "test-feed"
	const url = "https://example.com/feed"

	if err := addFeed(name, url); err != nil {
		t.Fatalf("addFeed returned error: %v", err)
	}

	if len(Config.Feeds) != 1 {
		t.Fatalf("expected 1 feed in memory, got %d", len(Config.Feeds))
	}

	if Config.Feeds[0].Name != name || Config.Feeds[0].URL != url {
		t.Fatalf("unexpected feed stored in memory: %+v", Config.Feeds[0])
	}

	cfgPath, err := configFilePath()
	if err != nil {
		t.Fatalf("configFilePath returned error: %v", err)
	}

	cfg, err := ini.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if got := cfg.Section("feeds").Key(name).String(); got != url {
		t.Fatalf("expected config to persist feed URL %s, got %s", url, got)
	}

	if err := addFeed(name, url); err == nil {
		t.Fatalf("expected adding duplicate feed to fail")
	}
}
