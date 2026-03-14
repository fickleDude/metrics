package repository

import (
	model "github.com/fickleDude/metrics.git/internal/model"
)

type MemStorageInterface interface {
	UpdateCount(name string, delta int64)
	UpdateGauge(name string, value float64)
	GetMetric(name string) *model.Metrics
	GetMetrics() []*model.Metrics
}
type MemStorage struct {
	storage []*model.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: []*model.Metrics{}}
}

func (s *MemStorage) UpdateCount(name string, delta int64) {
	for _, m := range s.storage {
		if m.ID == name {
			*m.Delta += delta
			return
		}
	}
	s.storage = append(s.storage, &model.Metrics{ID: name, MType: model.Counter, Delta: &delta, Value: nil})
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	for _, m := range s.storage {
		if m.ID == name {
			*m.Value = value
			return
		}
	}
	s.storage = append(s.storage, &model.Metrics{ID: name, MType: model.Gauge, Delta: nil, Value: &value})
}

func (s *MemStorage) GetMetric(name string) *model.Metrics {
	for _, m := range s.storage {
		if m.ID == name {
			return m
		}
	}
	return nil
}

func (s *MemStorage) GetMetrics() []*model.Metrics {
	return s.storage
}
