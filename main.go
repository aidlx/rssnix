package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	LoadConfig()
	log.Info("30 articles fetched from blah")
}
