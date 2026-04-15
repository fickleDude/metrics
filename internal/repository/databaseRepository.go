package repository

import (
	"database/sql"

	"github.com/fickleDude/metrics.git/internal/logger"
	model "github.com/fickleDude/metrics.git/internal/model"
)

type MemDatabaseStorage struct {
	MemStorage
	db *sql.DB
}

func NewMemDatabaseStorage(db *sql.DB) *MemDatabaseStorage {
	return &MemDatabaseStorage{MemStorage: *NewMemStorage(nil), db: db}
}

func (s *MemDatabaseStorage) UpdateCount(name string, delta *int64) {
	_, err := s.db.Exec(`INSERT INTO counter (id, delta) 
								VALUES ($1, $2)
								ON CONFLICT (id) DO UPDATE 
								SET delta = counter.delta + excluded.delta`, name, delta)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func (s *MemDatabaseStorage) UpdateGauge(name string, value *float64) {
	_, err := s.db.Exec(`INSERT INTO gauge (id, value) 
								VALUES ($1, $2)
								ON CONFLICT (id) DO UPDATE 
								SET value = excluded.value`, name, value)
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func (s *MemDatabaseStorage) GetMetric(name string, mType string) *model.Metrics {
	metric := model.Metrics{ID: name, MType: mType}
	var err error
	switch mType {
	case "gauge":
		row := s.db.QueryRow("select value from gauge where id = $1", name)
		err = row.Scan(&metric.Value)

	case "counter":
		row := s.db.QueryRow("select delta from counter where id = $1", name)
		err = row.Scan(&metric.Delta)
	}
	if err != nil {
		return nil
	}
	return &metric
}

func (s *MemDatabaseStorage) GetMetrics() []*model.Metrics {
	var metrics []*model.Metrics

	rows, err := s.db.Query("select id, delta from counter")
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		metric := model.Metrics{MType: model.Counter}
		err = rows.Scan(&metric.ID, &metric.Delta)
		if err != nil {
			return nil
		}

		metrics = append(metrics, &metric)
	}
	err = rows.Err()
	if err != nil {
		return nil
	}

	rows, err = s.db.Query("select id, value from gauge")
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		metric := model.Metrics{MType: model.Gauge}
		err = rows.Scan(&metric.ID, &metric.Value)
		if err != nil {
			return nil
		}

		metrics = append(metrics, &metric)
	}
	err = rows.Err()
	if err != nil {
		return nil
	}

	return metrics
}
