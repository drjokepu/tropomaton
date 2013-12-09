package main

import "fmt"
import "os"
import "time"

func main() {
	initConfig()
	mainLoopErrorHandler()
}

const tvTroperUrlPrefix = "http://tvtropes.org/pmwiki/pmwiki.php/"

func mainLoopErrorHandler() {
	for {
		err := mainLoop()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(5 * time.Second) // 0.5 seconds
		} else {
			return
		}
	}
}

func mainLoop() error {
	stop := false
	for !stop {
		nextLink, err := getNextLink()
		if err != nil {
			return err
		}

		var nextUrl string
		if nextLink == nil {
			hasAnyPagesAtAll, err := hasAnyPages()
			if err != nil {
				return err
			}

			if !hasAnyPagesAtAll {
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
