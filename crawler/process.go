package main

func process(url string) error {
	htmlChannel := make(chan string)
	errorChannel := make(chan error)
	go fetch(url, htmlChannel, errorChannel)

	select {
	case html := <-htmlChannel:
		return processHtml(url, html)
	case err := <-errorChannel:
		return err
	}

	return nil
}

func processHtml(url, contents string) error {
	pageWithLinks := parseHtml(url, contents)
	return pageWithLinks.save()
}
