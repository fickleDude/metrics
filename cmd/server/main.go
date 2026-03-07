package main

import (
	"net/http"

	"github.com/fickleDude/metrics.git/internal/handler"
)

func main() {

	//server
	handler := handler.MemStorageHandler{}
	http.HandleFunc(`/update/`, handler.UpdateHandler)

	err := http.ListenAndServe(`:8081`, nil)
	if err != nil {
		panic(err)
	}

}
