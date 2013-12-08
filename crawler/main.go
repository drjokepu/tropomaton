package main

import "time"

func main() {
	newDb, err := initdb()
	if err != nil {
		panic(err)
	}

	err = mainLoop(newDb)
	if err != nil {
		panic(err)
	}
}

const tvTroperUrlPrefix = "http://tvtropes.org/pmwiki/pmwiki.php/"

func mainLoop(newDb bool) error {
	stop := false
	for !stop {
		nextLink, err := getNextLink()
		if err != nil {
			return err
		}

		var nextUrl string
		if nextLink == nil {
			if newDb {
				nextUrl = "Main/HomePage"
			} else {
				stop = true
				break
			}
		} else {
			nextUrl = nextLink.href
		}

		err = process(tvTroperUrlPrefix + nextUrl)
		if err != nil {
			return err
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}
