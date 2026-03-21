package main

import (
	"net/http"

	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/handler"
	"github.com/fickleDude/metrics.git/internal/logger"
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
	// r.Get("/", middleware.RequestLogger(handler.GetMetricsHandler))
	// r.Get("/value/{type}/{name}", middleware.RequestLogger(handler.GetMetricHandler))
	// r.Post("/update", handler.UpdateMetricJsonHandler)
	r.Route("/", func(r chi.Router) {
		r.Post("/value/", handler.GetMetricJSONHandler)
		r.Post("/update/", handler.UpdateMetricJSONHandler)
	})
	// r.Post("/value/", handler.GetMetricJsonHandler)
	// r.Post("/update/{type}/{name}/{value}", middleware.RequestLogger(handler.UpdateMetricHandler))

	err := http.ListenAndServe(cfg.RunAddr(), r)
	if err != nil {
		panic(err)
	}

}
