package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/handler"
	"github.com/fickleDude/metrics.git/internal/logger"
	"github.com/fickleDude/metrics.git/internal/middleware"
	models "github.com/fickleDude/metrics.git/internal/model"
	"github.com/fickleDude/metrics.git/internal/repository"
	"github.com/fickleDude/metrics.git/internal/service"

	chi "github.com/go-chi/chi/v5"
)

func storeMetrics(serverAddress string, storeInterval int, filename string) {
	for {
		//create request
		target := fmt.Sprintf("http://%s/values", serverAddress)
		request, err := http.NewRequest(http.MethodGet, target, nil)
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		//get response
		client := http.Client{}
		response, err := client.Do(request)
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			logger.Log.Error(err.Error())
			return
		}
		//write to file
		os.WriteFile(filename, data, 0666)
		time.Sleep(time.Duration(storeInterval) * time.Second)
	}
}
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
	restore := cfg.Restore()
	metrics := []*models.Metrics{}
	if restore { //загружаем ранее сохранённые значения при старте сервера
		values, _ := os.ReadFile(cfg.FileStoragePath())
		reader := strings.NewReader(string(values))
		json.NewDecoder(reader).Decode(&metrics)

	}
	repository := repository.NewMemStorage(metrics)
	service := service.NewMemStorageService(repository)
	handler := handler.NewMemStorageHandler(service)

	//router
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Gzip)
	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetMetricsHTMLHandler)
		r.Get("/values", handler.GetMetricsHandler)
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
	go storeMetrics(cfg.RunAddr(), cfg.StoreInterval(), cfg.FileStoragePath())
	//start server
	err := http.ListenAndServe(cfg.RunAddr(), r)
	if err != nil {
		panic(err)
	}
}
