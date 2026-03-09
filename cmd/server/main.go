package main

import (
	"net/http"

	"github.com/fickleDude/metrics.git/internal/handler"
	"github.com/fickleDude/metrics.git/internal/repository"
	"github.com/fickleDude/metrics.git/internal/service"
	chi "github.com/go-chi/chi/v5"
)

func main() {

	//init
	repository := repository.NewMemStorage()
	service := service.NewMemStorageService(repository)
	handler := handler.NewMemStorageHandler(service)

	//router
	r := chi.NewRouter()
	r.Get("/", handler.GetMetricsHandler)
	r.Get("/value/{type}/{name}", handler.GetMetricHandler)
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetricHandler)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}

}
