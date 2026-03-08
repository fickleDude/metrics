package service

import (
	repository "github.com/fickleDude/metrics.git/internal/repository"
)

type MemStorageInterface interface {
	WriteCount(name string, value int64) error
	WriteGauge(name string, value float64) error
}

type MemStorageService struct {
	repository repository.MemStorageRepository
}

func (r *MemStorageService) WriteCount(name string, value int64) error {
	return r.repository.WriteCount(name, value)
}

func (r *MemStorageService) WriteGauge(name string, value int64) error {
	return r.repository.WriteGauge(name, value)
}
