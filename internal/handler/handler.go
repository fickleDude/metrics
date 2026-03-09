package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fickleDude/metrics.git/internal/service"
	"github.com/go-chi/chi/v5"
)

type MemStorageHandler struct {
	service service.MemStorageInterface
}

func NewMemStorageHandler(service service.MemStorageInterface) *MemStorageHandler {
	return &MemStorageHandler{service: service}
}

func (h *MemStorageHandler) UpdateMetricHandler(res http.ResponseWriter, req *http.Request) {
	memType := chi.URLParam(req, "type")
	memName := chi.URLParam(req, "name")
	memValue := chi.URLParam(req, "value")

	switch memType {
	case "counter":
		memValueInt, err := strconv.ParseInt(memValue, 10, 0)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		h.service.UpdateCount(memName, memValueInt)
		res.Write([]byte(fmt.Sprintf("metric %s updated. value = %d", memName, memValueInt)))

	case "gauge":
		memValueFloat, err := strconv.ParseFloat(memValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		h.service.UpdateGauge(memName, memValueFloat)
		res.Write([]byte(fmt.Sprintf("metric %s updated. value = %f", memName, memValueFloat)))
	default:
		res.WriteHeader(http.StatusBadRequest) //некорректный тип метрики
		return
	}
}

func (h *MemStorageHandler) GetMetricHandler(res http.ResponseWriter, req *http.Request) {
	memName := chi.URLParam(req, "name")
	memValue := h.service.GetMetric(memName)
	if memValue == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	res.Write([]byte(memValue))
}

func (h *MemStorageHandler) GetMetricsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	memValues := h.service.GetMetrics()
	body := fmt.Sprintf(`<!DOCTYPE html>
			<html lang="en">
			<head>
				<!-- Metadata goes here -->
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Page Title</title>
				<link rel="stylesheet" href="style.css">
			</head>
			<body>
				%s
			</body>
			</html>`, memValues)

	fmt.Fprint(res, body)
}
