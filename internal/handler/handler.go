package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fickleDude/metrics.git/internal/logger"
	models "github.com/fickleDude/metrics.git/internal/model"
	"github.com/fickleDude/metrics.git/internal/service"
	"github.com/go-chi/chi/v5"
)

type MemStorageHandler struct {
	service service.MemStorageInterface
}

func NewMemStorageHandler(service service.MemStorageInterface) *MemStorageHandler {
	return &MemStorageHandler{service: service}
}

func (h *MemStorageHandler) UpdateMetricJSONHandler(res http.ResponseWriter, req *http.Request) {
	//check content type
	if req.Header.Get("Content-Type") != "application/json" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	//decode request
	var metric models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		logger.Log.Error(err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	//update repository data
	switch metric.MType {
	case "counter":
		h.service.UpdateCount(metric.ID, metric.Delta)
	case "gauge":
		h.service.UpdateGauge(metric.ID, metric.Value)
	default:
		res.WriteHeader(http.StatusBadRequest) //некорректный тип метрики
		return
	}

}

func (h *MemStorageHandler) UpdateMetricHandler(res http.ResponseWriter, req *http.Request) {
	memType := chi.URLParam(req, "type")
	memName := chi.URLParam(req, "name")
	memValue := chi.URLParam(req, "value")

	res.Header().Set("Content-Type", "text/html")
	switch memType {
	case "counter":
		memValueInt, err := strconv.ParseInt(memValue, 10, 0)
		if err != nil {
			logger.Log.Error(err.Error())
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		h.service.UpdateCount(memName, &memValueInt)
		res.Write([]byte(fmt.Sprintf("metric %s updated. value = %d", memName, memValueInt)))

	case "gauge":
		memValueFloat, err := strconv.ParseFloat(memValue, 64)
		if err != nil {
			logger.Log.Error(err.Error())
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		h.service.UpdateGauge(memName, &memValueFloat)
		res.Write([]byte(fmt.Sprintf("metric %s updated. value = %f", memName, memValueFloat)))
	default:
		res.WriteHeader(http.StatusBadRequest) //некорректный тип метрики
		return
	}
}

func (h *MemStorageHandler) GetMetricJSONHandler(res http.ResponseWriter, req *http.Request) {
	//check content type
	if req.Header.Get("Content-Type") != "application/json" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	//decode request
	var metric models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		logger.Log.Error(err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	//get repository data
	repoMetric := h.service.GetMetric(metric.ID, metric.MType)
	if repoMetric == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	//encode response
	res.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(repoMetric); err != nil {
		logger.Log.Error(err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	buf.WriteTo(res)
}

func (h *MemStorageHandler) GetMetricHandler(res http.ResponseWriter, req *http.Request) {
	memName := chi.URLParam(req, "name")
	memType := chi.URLParam(req, "type")
	memValue := h.service.GetMetricValue(memName, memType)
	if memValue == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	res.Write([]byte(memValue))
}

func (h *MemStorageHandler) GetMetricsHTMLHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	memValues := h.service.GetMetricValues()
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

func (h *MemStorageHandler) GetMetricsHandler(res http.ResponseWriter, req *http.Request) {
	//get repository data
	repoMetric := h.service.GetMetrics()
	if repoMetric == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	//encode response
	res.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(repoMetric); err != nil {
		logger.Log.Error(err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	buf.WriteTo(res)
}
