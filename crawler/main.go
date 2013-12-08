package main

import "fmt"
import "os"
import "time"

func main() {
	newDb, err := initdb()
	if err != nil {
		panic(err)
	}

	mainLoopErrorHandler(newDb)
}

const tvTroperUrlPrefix = "http://tvtropes.org/pmwiki/pmwiki.php/"

func mainLoopErrorHandler(newDb bool) {
	stillNewDb := newDb
	for {
		err := mainLoop(stillNewDb)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			stillNewDb = false
		} else {
			return
		}
	}
}

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

		time.Sleep(500 * time.Millisecond) // 0.5 seconds
	}

	return nil
}
