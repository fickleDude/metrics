package repository

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/fickleDude/metrics.git/internal/logger"
	model "github.com/fickleDude/metrics.git/internal/model"
	models "github.com/fickleDude/metrics.git/internal/model"
)

type MemStorageInterface interface {
	UpdateCount(name string, delta *int64)
	UpdateGauge(name string, value *float64)
	GetMetric(name string, mType string) *model.Metrics
	GetMetrics() []*model.Metrics
	LoadFromFile(filename string) error
	LoadToFile(filename string) error
}
type MemStorage struct {
	storage []*model.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: []*model.Metrics{}}
}

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

// file
func (s *MemStorage) LoadFromFile(filename string) error {
	metrics := []*models.Metrics{}
	values, _ := os.ReadFile(filename)
	reader := strings.NewReader(string(values))
	if err := json.NewDecoder(reader).Decode(&metrics); err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	s.storage = metrics
	return nil
}

func (s *MemStorage) LoadToFile(filename string) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(s.storage); err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	os.WriteFile(filename, buf.Bytes(), 0666)
	return nil
}
