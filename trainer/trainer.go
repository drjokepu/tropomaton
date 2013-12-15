package main

import "fmt"
import "net/http"
import "os"

func main() {
	http.HandleFunc("/train", func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm();
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			http.Error(writer, "Internal Server Error", 500)
		}
		
		pageId := request.Form["pageId"][0]
		class := request.Form["class"][0]
		
		fmt.Println("Page:", pageId, "Class:", class)
	})
	
	http.ListenAndServe("localhost:8877", nil)
}