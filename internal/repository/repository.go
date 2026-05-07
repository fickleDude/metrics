package repository

import (
	model "github.com/fickleDude/metrics.git/internal/model"
)

type MemStorageInterface interface {
	UpdateCount(name string, delta *int64)
	UpdateGauge(name string, value *float64)
	GetMetric(name string, mType string) *model.Metrics
	GetMetrics() []*model.Metrics
	//GetDatabaseConnection() string
}
type MemStorage struct {
	storage []*model.Metrics
}

func NewMemStorage(data []*model.Metrics) *MemStorage {
	return &MemStorage{storage: data}
}

// func (s *MemStorage) GetDatabaseConnection() string {
// 	return s.databaseDns
// }

func (s *MemStorage) UpdateCount(name string, delta *int64) {
	for _, m := range s.storage {
		if m.ID == name {
			*m.Delta += *delta
			return
		}
	}
	s.storage = append(s.storage, &model.Metrics{ID: name, MType: model.Counter, Delta: delta, Value: nil})
}

func (s *MemStorage) UpdateGauge(name string, value *float64) {
	for _, m := range s.storage {
		if m.ID == name {
			m.Value = value
			return
		}
	}
	s.storage = append(s.storage, &model.Metrics{ID: name, MType: model.Gauge, Delta: nil, Value: value})
}

func (s *MemStorage) GetMetric(name string, mType string) *model.Metrics {
	for _, m := range s.storage {
		if m.MType == mType && m.ID == name {
			return m
		}
	}
	return nil
}

func (s *MemStorage) GetMetrics() []*model.Metrics {
	return s.storage
}
