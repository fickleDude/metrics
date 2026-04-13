package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fickleDude/metrics.git/internal/logger"
	models "github.com/fickleDude/metrics.git/internal/model"
	repository "github.com/fickleDude/metrics.git/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type MemStorageInterface interface {
	UpdateCount(name string, delta *int64)
	UpdateGauge(name string, value *float64)
	GetMetricValue(name string, mType string) string
	GetMetricValues() string
	GetMetric(name string, mType string) *models.Metrics
	GetMetrics() []*models.Metrics
	GetDbConnection() bool
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

func (r *MemStorageService) GetDbConnection() bool {
	db, err := sql.Open("pgx", r.repository.GetDatabaseConnection())
	if err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	return true
}
