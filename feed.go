package main

import (
	"os"
	"strconv"
	"sync"

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

func DeleteFeedFiles(name string) {
	os.RemoveAll(Config.FeedDirectory + "/" + name)
	os.Mkdir(Config.FeedDirectory+"/"+name, 0644)
}

func UpdateFeed(name string) {
	log.Info("Updating feed '" + name + "'")
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(Config.Feeds[slices.IndexFunc(Config.Feeds, func(f Feed) bool { return f.Name == name })].URL)
	DeleteFeedFiles(name)
	for _, item := range feed.Items {
		file, err := os.Create(Config.FeedDirectory + "/" + name + "/" + item.Title)
		if err != nil {
			log.Error("Failed to create a file for article titled '" + item.Title + "'")
			continue
		}
		defer file.Close()
		_, err = file.WriteString(item.Content)
		if err != nil {
			log.Error("Failed to write content to a file for article titled '" + item.Title + "'")
			continue
		}
	}
	log.Info(strconv.Itoa(len(feed.Items)) + " articles fetched from feed '" + name + "'")
	if isAllUpdate {
		wg.Done()
	}
}

func UpdateAllFeeds() {
	isAllUpdate = true
	for _, feed := range Config.Feeds {
		wg.Add(1)
		go UpdateFeed(feed.Name)
	}
	wg.Wait()
}
