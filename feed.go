package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type Feed struct {
	Name string
	URL  string
}

type FeedUpdateResult struct {
	Name       string
	Downloaded int
	Skipped    int
	Total      int
}

const newArticleDirectory = "new"
const maxFileNameLength = 255

func truncateString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.ValidString(s[:n]) {
		n--
	}
	if n <= 0 {
		return ""
	}
	return s[:n]
}

func safeArticleName(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return ""
	}

	sanitized := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return -1
		default:
			if r < 32 {
				return -1
			}
			return r
		}
	}, trimmed)

	return sanitized
}

func InitialiseNewArticleDirectory() error {
	if err := DeleteFeedFiles(newArticleDirectory); err != nil {
		return fmt.Errorf("clean new article directory: %w", err)
	}
	return os.MkdirAll(filepath.Join(Config.FeedDirectory, newArticleDirectory), 0o755)
}

func DeleteFeedFiles(name string) error {
	return os.RemoveAll(filepath.Join(Config.FeedDirectory, name))
}

func UpdateFeed(name string, deleteFiles bool) (FeedUpdateResult, error) {
	result := FeedUpdateResult{Name: name}

	feedConfig, ok := Config.FeedByName(name)
	if !ok {
		return result, fmt.Errorf("feed %q not found", name)
	}

	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(feedConfig.URL)
	if err != nil {
		return result, fmt.Errorf("fetch feed %q: %w", name, err)
	}

	result.Total = len(feed.Items)

	if deleteFiles {
		if err := DeleteFeedFiles(name); err != nil {
			log.WithError(err).Errorf("Failed to delete existing articles for feed '%s'", name)
		}
	}

	feedDir := filepath.Join(Config.FeedDirectory, name)
	if err := os.MkdirAll(feedDir, 0o755); err != nil {
		return result, fmt.Errorf("ensure feed directory for %q: %w", name, err)
	}

	for _, item := range feed.Items {
		articleName := truncateString(safeArticleName(item.Title), maxFileNameLength)
		if articleName == "" {
			log.WithField("feed", name).Warn("Skipping item with empty or invalid title")
			result.Skipped++
			continue
		}

		articlePath := filepath.Join(feedDir, articleName)
		if _, err := os.Stat(articlePath); err == nil {
			log.Debugf("Article %s already exists - skipping download", articlePath)
			result.Skipped++
			continue
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.WithError(err).Warnf("Unable to check if article exists: %s", articlePath)
			result.Skipped++
			continue
		}

		file, err := os.Create(articlePath)
		if err != nil {
			log.WithError(err).Errorf("Failed to create file for article titled '%s'", item.Title)
			result.Skipped++
			continue
		}

		published := item.Published
		if published == "" && item.PublishedParsed != nil {
			published = item.PublishedParsed.Format(time.RFC3339)
		}

		var builder strings.Builder
		builder.WriteString(item.Description)
		builder.WriteByte('\n')
		builder.WriteString(item.Link)
		builder.WriteByte('\n')
		builder.WriteString(published)
		builder.WriteByte('\n')
		builder.WriteString(item.Content)

		if _, err := file.WriteString(builder.String()); err != nil {
			log.WithError(err).Errorf("Failed to write content for article titled '%s'", item.Title)
			file.Close()
			os.Remove(articlePath)
			result.Skipped++
			continue
		}

		if err := file.Close(); err != nil {
			log.WithError(err).Warnf("Failed to close file for article titled '%s'", item.Title)
		}

		result.Downloaded++

		newLinkPath := filepath.Join(Config.FeedDirectory, newArticleDirectory, articleName)
		if err := os.Remove(newLinkPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.WithError(err).Warnf("Failed to remove existing symlink for article %s", newLinkPath)
		}
		if err := os.Symlink(articlePath, newLinkPath); err != nil {
			log.WithError(err).Warnf("Could not create symlink for newly downloaded article %s", articlePath)
		}
	}

	log.Infof("%d articles fetched from feed '%s' (%d already seen, %d total in feed)", result.Downloaded, name, result.Skipped, result.Total)

	return result, nil
}

func UpdateAllFeeds(deleteFiles bool) []FeedUpdateResult {
	results := make([]FeedUpdateResult, 0, len(Config.Feeds))
	if len(Config.Feeds) == 0 {
		return results
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, feed := range Config.Feeds {
		feedName := feed.Name
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := UpdateFeed(feedName, deleteFiles)
			if err != nil {
				log.Error(err)
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return results
}
