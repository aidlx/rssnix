package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTruncateString(t *testing.T) {
	names := []string{
		"我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我",  // 255 x Chinese wo3 (我)
		"我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我", // 256 x Chinese wo3 (我)
		"我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我", // 32 x Chinese wo3 (我)
		"short",
	}

	for _, name := range names {
		shortened := truncateString(name, maxFileNameLength)
		if len(name) <= maxFileNameLength {
			if name != shortened {
				t.Errorf("expected %q to remain unchanged, got %q", name, shortened)
			}
		} else {
			if len([]byte(shortened)) > maxFileNameLength {
				t.Errorf("expected truncated string to be at most %d bytes, got %d", maxFileNameLength, len([]byte(shortened)))
			}
		}
	}

	if got := truncateString("abc", 0); got != "" {
		t.Errorf("expected empty string when n == 0, got %q", got)
	}
}

func TestSafeArticleName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  hello world  ", "hello world"},
		{"question?mark", "questionmark"},
		{"slash/path", "slashpath"},
		{"colon:name", "colonname"},
		{"|pipes|", "pipes"},
		{"\nnew\nlines", "newlines"},
		{"", ""},
	}

	for _, tc := range tests {
		if got := safeArticleName(tc.input); got != tc.want {
			t.Errorf("safeArticleName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestUpdateFeedCreatesArticles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(configEnvVar, "")

	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	const articleTitle = "Example Article / 1?"
	const articleContent = `<rss version="2.0"><channel><title>Test Feed</title><item><title>` + articleTitle + `</title><link>https://example.com/article</link><description>Description</description><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate><content:encoded xmlns:content="http://purl.org/rss/1.0/modules/content/">Full content</content:encoded></item></channel></rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(articleContent))
	}))
	t.Cleanup(server.Close)

	Config.Feeds = []Feed{{Name: "test-feed", URL: server.URL}}

	if err := InitialiseNewArticleDirectory(); err != nil {
		t.Fatalf("InitialiseNewArticleDirectory returned error: %v", err)
	}

	result, err := UpdateFeed("test-feed", true)
	if err != nil {
		t.Fatalf("UpdateFeed returned error: %v", err)
	}

	if result.Downloaded != 1 {
		t.Fatalf("expected 1 downloaded article, got %d", result.Downloaded)
	}

	sanitized := truncateString(safeArticleName(articleTitle), maxFileNameLength)
	articlePath := filepath.Join(Config.FeedDirectory, "test-feed", sanitized)
	if _, err := os.Stat(articlePath); err != nil {
		t.Fatalf("expected article file to exist at %s: %v", articlePath, err)
	}

	newLink := filepath.Join(Config.FeedDirectory, newArticleDirectory, sanitized)
	target, err := os.Readlink(newLink)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", newLink, err)
	}

	if target != articlePath {
		t.Fatalf("expected symlink target %s, got %s", articlePath, target)
	}
}

func TestUpdateFeedMissingFeed(t *testing.T) {
	orig := Config
	t.Cleanup(func() { Config = orig })
	Config = Configuration{}
	if _, err := UpdateFeed("missing", false); err == nil {
		t.Fatalf("expected error when updating missing feed")
	}
}
