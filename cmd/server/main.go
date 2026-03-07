package main

import (
	"fmt"
	"net/http"

	"github.com/fickleDude/metrics.git/internal/handler"
)

func main() {

	//server
	handler := handler.MemStorageHandler{}
	http.HandleFunc(`/update/`, handler.UpdateHandler)

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("ready to serve on localhost:8080")
	}

}
