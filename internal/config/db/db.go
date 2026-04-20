package db

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	instance *sql.DB
	once     sync.Once
)

func GetDBConnection() *sql.DB {
	once.Do(func() {
		dataSourceName := config.GetConfig(config.Server).DatabaseDNS()
		db, err := sql.Open("pgx", dataSourceName)
		if err != nil {
			panic(err.Error())
		}
		instance = db
	})
	return instance
}

func CloseDBConnection() error {
	return instance.Close()
}

func TestDBConnection() bool {
	if instance == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := instance.PingContext(ctx); err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	return true
}
