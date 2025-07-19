package main

import (
	log "github.com/sirupsen/logrus"
)

type Feed struct {
	Name    string
	Address string
}

func UpdateAllFeeds() {
	log.Info("All feeds updated")
}

func UpdateFeed(feed string) {
	log.Info(feed + " updated")
}
