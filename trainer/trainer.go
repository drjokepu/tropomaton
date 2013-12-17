package main

import "flag"
import "fmt"
import "net/http"
import "os"
import "strconv"

func main() {
	initConfig()

	statsFlag := flag.Bool("stats", false, "Display stats.")
	trainFlag := flag.Bool("train", false, "Listen to training request.")
	guessFlag := flag.Int("guess", -1, "Guess the class of the page with the given id.")
	flag.Parse()

	switch {
	case *statsFlag:
		printStats()
	case *trainFlag:
		listen()
	case *guessFlag >= 0:
		printGuess(*guessFlag)
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

		err = trainPage(pageId, class)
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

func printGuess(pageId int) {
	guessedClass, err := guessPageClass(pageId)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to guess class:", err.Error())
		return
	}
	fmt.Println(guessedClass, "-", getClassName(guessedClass))
}

func printStats() {
	classifier, err := loadClassifier()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot load classifier:", err.Error())
	}

	fmt.Println(classifier.Learned(), "learned")
	fmt.Println(classifier.Seen(), "seen")
	wordCounts := classifier.WordCount()
	fmt.Println(wordCounts[tropeIndex], "trope")
	fmt.Println(wordCounts[notTropeIndex], "not trope")
	fmt.Println(wordCounts[workIndex], "work")
	fmt.Println(wordCounts[notWorkIndex], "not work")
}
