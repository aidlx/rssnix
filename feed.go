package main

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Feed struct {
	Name string
	URL  string
}

var wg sync.WaitGroup
var isAllUpdate bool
const maxFileNameLength = 255

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for !utf8.ValidString(s[:n]) {
		n--
	}
	return s[:n]
}

func InitialiseNewArticleDirectory() {
	DeleteFeedFiles(newArticleDirectory)
	os.MkdirAll(Config.FeedDirectory+"/"+newArticleDirectory, 0755)
}

func DeleteFeedFiles(name string) {
	os.RemoveAll(Config.FeedDirectory + "/" + name)
}

func UpdateFeed(name string, deleteFiles bool) {
	log.Info("Updating feed '" + name + "'")
	fp := gofeed.NewParser()
	downloadCount := 0
	skipCount := 0
	feed, err := fp.ParseURL(Config.Feeds[slices.IndexFunc(Config.Feeds, func(f Feed) bool { return f.Name == name })].URL)
	if err != nil {
		log.Error("Failed to fetch the feed '" + name + "'")
		if isAllUpdate {
			wg.Done()
		}
		return
	}
	if deleteFiles {
		DeleteFeedFiles(name)
	}
	os.MkdirAll(Config.FeedDirectory+"/"+name, 0777)
	for _, item := range feed.Items {

		articlePath := Config.FeedDirectory + "/" + name + "/" + truncateString(strings.ReplaceAll(item.Title, "/", ""), maxFileNameLength)
		if _, err := os.Stat(articlePath); err == nil {
			log.Info("Article " + articlePath + " already exists - skipping download")
			skipCount++
			continue
		}
		file, err := os.Create(articlePath)
		if err != nil {
			log.Error("Failed to create a file for article titled '" + item.Title + "'")
			continue
		}
		defer file.Close()
		_, err = file.WriteString(item.Description + "\n" + item.Link + "\n" + item.Published + "\n" + item.Content)
		if err != nil {
			log.Error("Failed to write content to a file for article titled '" + item.Title + "'")
			continue
		}
		downloadCount++
	}
	log.Info(strconv.Itoa(downloadCount) + " articles fetched from feed '" + name + "' (" + strconv.Itoa(skipCount) + " already seen, " + strconv.Itoa(len(feed.Items)) + " total in feed)")
	if isAllUpdate {
		wg.Done()
	}
}

func UpdateAllFeeds(deleteFiles bool) {
	isAllUpdate = true
	for _, feed := range Config.Feeds {
		wg.Add(1)
		go UpdateFeed(feed.Name, deleteFiles)
	}
	wg.Wait()
}
