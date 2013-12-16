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
			fmt.Fprintln(os.Stderr, err.Error())
			http.Error(writer, "Internal Server Error", 500)
		}

		pageId, err := strconv.Atoi(request.Form["pageId"][0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			http.Error(writer, "Internal Server Error", 500)
		}

		class, err := strconv.Atoi(request.Form["class"][0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			http.Error(writer, "Internal Server Error", 500)
		}

		train(pageId, class)
	})

	http.ListenAndServe("localhost:8877", nil)
}
