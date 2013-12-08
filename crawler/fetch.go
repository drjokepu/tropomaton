package main

import "io/ioutil"
import "fmt"
import "net/http"

func fetch(url string, outputChannel chan string, errorChannel chan error) {
	fmt.Println("  fetching:", url)

	client := new(http.Client)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errorChannel <- err
		return
	}

	req.Header.Set("User-Agent", "TV Tropes Crawler - Contact: tomi.czinege@gmail.com")

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			outputChannel <- string(body)
		} else {
			errorChannel <- err
		}
	} else {
		errorChannel <- err
	}
}
