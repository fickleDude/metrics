package main

import (
	"net/http"
	"time"

	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/handler"
	"github.com/fickleDude/metrics.git/internal/logger"
	"github.com/fickleDude/metrics.git/internal/middleware"
	"github.com/fickleDude/metrics.git/internal/repository"
	s "github.com/fickleDude/metrics.git/internal/service"

	chi "github.com/go-chi/chi/v5"
)

func main() {

	//config
	cfg := config.NewConfig()
	cfg.ParseFlags("server")
	cfg.ParseEnv("server")

	//init logger
	logLevel := "info"
	if err := logger.Initialize(logLevel); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	//init
	repository := repository.NewMemStorage()
	if cfg.Restore() {
		repository.LoadFromFile(cfg.FileStoragePath())
	}

	var service s.MemStorageInterface
	service = s.NewMemStorageService(repository)
	if cfg.StoreInterval() == 0 {
		service = s.NewMemStorageSyncService(*s.NewMemStorageService(repository), cfg.FileStoragePath())
	}
	handler := handler.NewMemStorageHandler(service)

	//router
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Gzip)
	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetMetricsHTMLHandler)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handler.GetMetricJSONHandler)
			r.Get("/{type}/{name}", handler.GetMetricHandler)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handler.UpdateMetricJSONHandler)
			r.Post("/{type}/{name}/{value}", handler.UpdateMetricHandler)
		})
	})

	//store metrics to file
	if cfg.StoreInterval() > 0 {
		ticker := time.NewTicker(time.Duration(cfg.StoreInterval()) * time.Second)
		go func() {
			for {
				repository.LoadToFile(cfg.FileStoragePath())
				<-ticker.C
			}
		}()
	}

	//start server
	err := http.ListenAndServe(cfg.RunAddr(), r)
	if err != nil {
		panic(err)
	}
}
