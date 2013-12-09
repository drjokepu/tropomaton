package main

import "encoding/json"
import "io/ioutil"

var sharedConfig map[string]interface{}
const databaseConnectionStringConfigKey = "databaseConnectionString"

func initConfig() {
	sharedConfig = loadConfig()
}

func loadConfig() map[string]interface{} {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		panic(err)
	}

	return result
}
