package main

import (
	"net/http"

	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/config/db"
	"github.com/fickleDude/metrics.git/internal/handler"
	"github.com/fickleDude/metrics.git/internal/logger"
	"github.com/fickleDude/metrics.git/internal/middleware"
	"github.com/fickleDude/metrics.git/internal/repository"
	s "github.com/fickleDude/metrics.git/internal/service"

	chi "github.com/go-chi/chi/v5"
)

func main() {

	//config
	cfg := config.GetConfig(config.Server)

	//init logger
	logLevel := "info"
	if err := logger.Initialize(logLevel); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	//init
	var memRepository repository.MemStorageInterface
	if cfg.DatabaseDns() != "" {
		defer db.CloseDbConnection()
		memRepository = repository.NewMemDatabaseStorage(db.GetDbConnection())
	} else if cfg.FileStoragePath() != "" {
		memRepository = repository.NewMemFileStorage(cfg.FileStoragePath(), cfg.Restore(), cfg.StoreInterval())
		if cfg.StoreInterval() > 0 {
			go memRepository.(*repository.MemFileStorage).SyncToFile()
		}
	} else {
		memRepository = repository.NewMemStorage(nil)
	}
	service := s.NewMemStorageService(memRepository)
	handler := handler.NewMemStorageHandler(service)

	//router
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Gzip)
	r.Route("/", func(r chi.Router) {
		r.Get("/ping", handler.GetDbConnectionHandler)
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

	//start server
	err := http.ListenAndServe(cfg.RunAddr(), r)
	if err != nil {
		panic(err)
	}
}
