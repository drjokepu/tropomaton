package main

import "database/sql"
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

		fmt.Println("Page:", pageId, "Class:", class)
		train(pageId, class)
	})

	http.ListenAndServe("localhost:8877", nil)
}

func train(pageId, class int) error {
	err := run(func(tx *sql.Tx) error {
		page, err := getPage(pageId, tx)
		if err != nil {
			return err
		}
		page = page
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
