package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fickleDude/metrics.git/internal/logger"
	model "github.com/fickleDude/metrics.git/internal/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/fickleDude/metrics.git/migrations"
)

type MemDatabaseStorage struct {
	MemStorage
	db *sql.DB
}

func NewMemDatabaseStorage(db *sql.DB) *MemDatabaseStorage {
	repository := &MemDatabaseStorage{MemStorage: *NewMemStorage(nil), db: db}
	err := repository.initDatabase()
	if err != nil {
		logger.Log.Error(err.Error())
	}
	return repository
}

func (s *MemDatabaseStorage) initDatabase() error {
	m := migrations.GetMigrator()
	if m != nil {
		return m.MigrateUp()
	}
	return fmt.Errorf("migration unavailable")
}

func isRetriable(pgErr *pgconn.PgError) bool {
	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return true

	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return true

	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return true
	}

	// Можно добавить более конкретные проверки с использованием констант pgerrcode
	switch pgErr.Code {
	// Класс 22 - Ошибки данных
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return false

	// Класс 23 - Нарушение ограничений целостности
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return false

	// Класс 42 - Синтаксические ошибки
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return false
	}

	// По умолчанию считаем ошибку неповторяемой
	return false
}

func (s *MemDatabaseStorage) executeWithRetry(query string, args ...any) error {
	maxRetries := 3
	retryDelay := []int{1, 3, 5}
	var lastError error

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, lastError = s.db.Exec(query, args...)
		if lastError == nil {
			return nil
		}
		//определяем тип ошибки
		//конвертируем в pgconn.PgError
		var pgErr *pgconn.PgError
		if errors.As(lastError, &pgErr) {
			// класс 08 - Ошибки соединения
			if isRetriable(pgErr) {
				time.Sleep(time.Duration(retryDelay[attempt]) * time.Second)
				continue
			}
		}
		return lastError
	}
	return fmt.Errorf("after %d attempts, last error: %w", maxRetries, lastError)
}

func (s *MemDatabaseStorage) queryWithRetry(query string, args ...any) (*sql.Rows, error) {
	maxRetries := 3
	retryDelay := []int{1, 3, 5}
	var lastError error

	for attempt := range maxRetries {
		rows, lastError := s.db.Query(query, args...)
		if lastError == nil {
			return rows, nil
		}
		//определяем тип ошибки
		//конвертируем в pgconn.PgError
		var pgErr *pgconn.PgError
		if errors.As(lastError, &pgErr) {
			// класс 08 - Ошибки соединения
			if isRetriable(pgErr) {
				time.Sleep(time.Duration(retryDelay[attempt]) * time.Second)
				continue
			}
		} else {
			return nil, lastError
		}
	}
	return nil, fmt.Errorf("after %d attempts, last error: %w", maxRetries, lastError)
}

func (s *MemDatabaseStorage) UpdateCount(name string, delta *int64) {
	err := s.executeWithRetry(`INSERT INTO counter (id, delta) 
								VALUES ($1, $2)
								ON CONFLICT (id) DO UPDATE 
								SET delta = counter.delta + excluded.delta`, name, delta)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

func (s *MemDatabaseStorage) UpdateGauge(name string, value *float64) {
	err := s.executeWithRetry(`INSERT INTO gauge (id, value) 
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

	rows, err := s.queryWithRetry("select id, delta from counter")
	if err != nil {
		logger.Log.Error(err.Error())
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

	rows, err = s.queryWithRetry("select id, value from gauge")
	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}
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
