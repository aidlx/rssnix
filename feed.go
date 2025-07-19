package main

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Feed struct {
	Name    string
	Address string
}

func UpdateFeed(feed string) {
	log.Info("Updating feed '" + feed + "'")
	log.Info(strconv.Itoa(1) + " articles fetched from feed '" + feed + "'")
}

func UpdateAllFeeds() {
	for _, feed := range Config.Feeds {
		UpdateFeed(feed.Name)
	}
}
