package main

import "encoding/json"
import "io/ioutil"

var sharedConfig *config

type config struct {
	databaseConnectionString string
}

func initConfig() {
	sharedConfig = loadConfig()
}

func loadConfig() *config {
	bytes, err := ioutil.ReadFile("crawler.config.json")
	if err != nil {
		panic(err)
	}

	var result config
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		panic(err)
	}

	return &result
}
