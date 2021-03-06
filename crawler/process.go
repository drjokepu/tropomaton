package main

func process(url string, getAllLinks bool) error {
	htmlChannel := make(chan string)
	errorChannel := make(chan error)
	go fetch(url, htmlChannel, errorChannel)

	select {
	case html := <-htmlChannel:
		return processHtml(url, html, getAllLinks)
	case err := <-errorChannel:
		return err
	}

	return nil
}

func processHtml(url, contents string, getAllLinks bool) error {
	pageWithLinks := parseHtml(url, contents, getAllLinks)
	return pageWithLinks.save()
}
