package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/fickleDude/metrics.git/internal/service"
)

type MemStorageHandler struct {
	service service.MemStorageInterface
}

func (h *MemStorageHandler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data := strings.Split(req.URL.Path, "/")
	if len(data) < 5 {
		res.WriteHeader(http.StatusNotFound) //имя метрики не указано
		return
	}

	metricType := data[2]
	metricName := data[3]
	metricValue := data[4]

	switch metricType {
	case "counter":
		metricValueInt, err := strconv.ParseInt(metricValue, 10, 0)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		//h.service.WriteCount(metricName, metricValueInt)
		res.Write([]byte(fmt.Sprintf("metric %s updated. value = %d", metricName, metricValueInt)))
	case "gauger":
		metricValueFloat, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest) //некорректное значение
			return
		}
		//h.service.WriteGauge(metricName, metricValueFloat)
		res.Write([]byte(fmt.Sprintf("metric %s updated. value = %f", metricName, metricValueFloat)))
	default:
		res.WriteHeader(http.StatusBadRequest) //некорректный тип метрики
		return
	}
}
