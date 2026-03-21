package main

import (
	"net/http"

	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/handler"
	"github.com/fickleDude/metrics.git/internal/logger"
	"github.com/fickleDude/metrics.git/internal/middleware"
	"github.com/fickleDude/metrics.git/internal/repository"
	"github.com/fickleDude/metrics.git/internal/service"

	chi "github.com/go-chi/chi/v5"
)

func main() {

	//config
	cfg := config.NewConfig()
	cfg.ParseFlags()
	cfg.ParseEnv()

	//init logger
	logLevel := "info"
	if err := logger.Initialize(logLevel); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	//init
	repository := repository.NewMemStorage()
	service := service.NewMemStorageService(repository)
	handler := handler.NewMemStorageHandler(service)

	//router
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetMetricsHandler)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handler.GetMetricJSONHandler)
			r.Get("/{type}/{name}", handler.GetMetricHandler)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handler.UpdateMetricJSONHandler)
			r.Post("/{type}/{name}/{value}", handler.UpdateMetricHandler)
		})
	})
	//start server
	err := http.ListenAndServe(cfg.RunAddr(), r)
	if err != nil {
		panic(err)
	}

}
