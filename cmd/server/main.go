package main

import (
	"net/http"

	"github.com/fickleDude/metrics.git/internal/handler"
)

func main() {
	handler := handler.MemStorageHandler{}
	http.HandleFunc(`/update/`, handler.UpdateHandler)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
