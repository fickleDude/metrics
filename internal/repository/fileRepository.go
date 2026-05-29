package repository

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/fickleDude/metrics.git/internal/logger"
	model "github.com/fickleDude/metrics.git/internal/model"
)

type MemFileStorage struct {
	MemStorage
	filename      string
	storeInterval int
}

func NewMemFileStorage(filename string, restore bool, storeInterval int) *MemFileStorage {
	var data []*model.Metrics
	if restore {
		data = loadFromFile(filename)
	}
	return &MemFileStorage{MemStorage: *NewMemStorage(data), filename: filename, storeInterval: storeInterval}
}

func (s *MemFileStorage) UpdateCount(name string, delta *int64) {
	for _, m := range s.storage {
		if m.ID == name {
			*m.Delta += *delta
			return
		}
	}
	s.storage = append(s.storage, &model.Metrics{ID: name, MType: model.Counter, Delta: delta, Value: nil})

	//sync changes
	if s.storeInterval == 0 {
		s.loadToFile()
	}
}

func (s *MemFileStorage) UpdateGauge(name string, value *float64) {
	for _, m := range s.storage {
		if m.ID == name {
			m.Value = value
			return
		}
	}
	s.storage = append(s.storage, &model.Metrics{ID: name, MType: model.Gauge, Delta: nil, Value: value})
	//sync changes
	if s.storeInterval == 0 {
		s.loadToFile()
	}
}

func (s *MemFileStorage) SyncToFile() {
	ticker := time.NewTicker(time.Duration(s.storeInterval) * time.Second)
	for {
		s.loadToFile()
		<-ticker.C
	}
}

// file
func loadFromFile(filename string) []*model.Metrics {
	metrics := []*model.Metrics{}
	values, _ := os.ReadFile(filename)
	reader := strings.NewReader(string(values))
	if err := json.NewDecoder(reader).Decode(&metrics); err != nil {
		logger.Log.Error(err.Error())
		return nil
	}
	return metrics
}

func (s *MemFileStorage) loadToFile() error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(s.storage); err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	os.WriteFile(s.filename, buf.Bytes(), 0666)
	return nil
}
