package service

import (
	"fmt"

	models "github.com/fickleDude/metrics.git/internal/model"
	repository "github.com/fickleDude/metrics.git/internal/repository"
)

type MemStorageInterface interface {
	UpdateCount(name string, delta *int64)
	UpdateGauge(name string, value *float64)
	GetMetricValue(name string, mType string) string
	GetMetricValues() string
	GetMetric(name string, mType string) *models.Metrics
	GetMetrics() []*models.Metrics
}

type MemStorageService struct {
	repository repository.MemStorageInterface
}

func NewMemStorageService(repository repository.MemStorageInterface) *MemStorageService {
	return &MemStorageService{repository: repository}
}
func (r *MemStorageService) UpdateCount(name string, delta *int64) {
	r.repository.UpdateCount(name, delta)
}

func (r *MemStorageService) UpdateGauge(name string, value *float64) {
	r.repository.UpdateGauge(name, value)
}

func (r *MemStorageService) GetMetricValue(name string, mType string) string {
	metric := r.repository.GetMetric(name, mType)
	if metric == nil {
		return ""
	}
	return metric.GetValue()
}

func (r *MemStorageService) GetMetricValues() string {
	metrics := r.repository.GetMetrics()
	var result string
	for _, m := range metrics {
		result += fmt.Sprintf("<p>%s : %s</p>", m.ID, m.GetValue())
	}
	return result
}

func (r *MemStorageService) GetMetric(name string, mType string) *models.Metrics {
	return r.repository.GetMetric(name, mType)
}

func (r *MemStorageService) GetMetrics() []*models.Metrics {
	return r.repository.GetMetrics()
}
