package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fickleDude/metrics.git/internal/config/db"
	"github.com/fickleDude/metrics.git/internal/helpers"
	"github.com/fickleDude/metrics.git/internal/logger"
	models "github.com/fickleDude/metrics.git/internal/model"
	"github.com/fickleDude/metrics.git/internal/service"
	"github.com/go-chi/chi/v5"
)

type MemStorageHandler struct {
	service service.MemStorageInterface
	signer  *helpers.Signer
}

func NewMemStorageHandler(service service.MemStorageInterface, signer *helpers.Signer) *MemStorageHandler {
	return &MemStorageHandler{service: service, signer: signer}
}

func (h *MemStorageHandler) UpdateMetricJSONHandler(res http.ResponseWriter, req *http.Request) {
	//check content type
	if req.Header.Get("Content-Type") != "application/json" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	if !h.signer.VerifyRequest(req) {
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

func (h *MemStorageHandler) UpdateMetricsJSONHandler(res http.ResponseWriter, req *http.Request) {
	//check content type
	if req.Header.Get("Content-Type") != "application/json" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	//check sign
	if !h.signer.VerifyRequest(req) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	//decode request
	var metrics []models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		logger.Log.Error(err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, metric := range metrics {
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

}

func (h *MemStorageHandler) UpdateMetricHandler(res http.ResponseWriter, req *http.Request) {
	if !h.signer.VerifyRequest(req) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

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

	case "gauge":
		memValueFloat, err := strconv.ParseFloat(memValue, 64)
		if err != nil {
			logger.Log.Error(err.Error())
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		h.service.UpdateGauge(memName, &memValueFloat)

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
	if !h.signer.VerifyRequest(req) {
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
	//sign
	h.signer.SignResponse(buf.Bytes(), res)
	//write response
	buf.WriteTo(res)
}

func (h *MemStorageHandler) GetMetricHandler(res http.ResponseWriter, req *http.Request) {
	if !h.signer.VerifyRequest(req) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	memName := chi.URLParam(req, "name")
	memType := chi.URLParam(req, "type")
	memValue := h.service.GetMetricValue(memName, memType)
	if memValue == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	//sign
	h.signer.SignResponse([]byte(memValue), res)
	//write response
	res.Write([]byte(memValue))
}

func (h *MemStorageHandler) GetMetricsHTMLHandler(res http.ResponseWriter, req *http.Request) {
	if !h.signer.VerifyRequest(req) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "text/html")
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
	//sign
	h.signer.SignResponse([]byte(body), res)
	//write response
	fmt.Fprint(res, body)
}

func (h *MemStorageHandler) GetDbConnectionHandler(res http.ResponseWriter, req *http.Request) {
	ok := db.TestDBConnection()
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
