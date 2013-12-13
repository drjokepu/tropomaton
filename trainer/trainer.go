package main

import "fmt"
import "net/http"

type trainerRequestHandler struct {}

func main() {
	handler := &trainerRequestHandler { }
	http.ListenAndServe("localhost:8877", handler)
}

func (handler *trainerRequestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("incoming request")
}