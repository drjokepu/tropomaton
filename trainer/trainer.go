package main

import "fmt"
import "net/http"
import "os"
import "strconv"

func main() {
	initConfig()

	http.HandleFunc("/train", func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to parse form.")
			reportError(writer, err)
			return
		}

		pageId, err := strconv.Atoi(request.Form["pageId"][0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to parse page id.")
			reportError(writer, err)
			return
		}

		class, err := strconv.Atoi(request.Form["class"][0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to parse class.")
			fmt.Fprintln(os.Stderr, err.Error())
			http.Error(writer, "Internal Server Error", 500)
		}

		err = train(pageId, class)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to train data.")
			reportError(writer, err)
			return
		}
	})

	http.ListenAndServe("localhost:8877", nil)
}

func reportError(writer http.ResponseWriter, err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	http.Error(writer, "Internal Server Error", 500)
}
