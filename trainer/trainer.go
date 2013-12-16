package main

import "flag"
import "fmt"
import "net/http"
import "os"
import "strconv"

func main() {
	initConfig()

	statsFlag := flag.Bool("stats", false, "Display stats")
	flag.Parse()
	
	switch {
	case *statsFlag:
		printStats();
	default:
		listen();
	}
}

func listen() {
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

func printStats() {
	classifier, err := loadClassifier()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot load classifier:", err.Error())
	}
	
	fmt.Println(classifier.Learned(), "learned")
	fmt.Println(classifier.Seen(), "seen")
	wordCounts := classifier.WordCount()
	fmt.Println(wordCounts[0], "trope")
	fmt.Println(wordCounts[1], "not trope")
	fmt.Println(wordCounts[2], "work")
	fmt.Println(wordCounts[3], "not work")
}
