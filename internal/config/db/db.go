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

type Database struct {
	db *sql.DB
}

var (
	instance *Database
	once     sync.Once
)

func GetDbConnection() *sql.DB {
	once.Do(func() {
		db, err := sql.Open("pgx", config.GetConfig().DatabaseDns())
		if err != nil {
			panic(err.Error())
		}
		instance = &Database{db: db}
	})
	return instance.db
}

func CloseDbConnection() error {
	return instance.db.Close()
}

func TestConnection() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := instance.db.PingContext(ctx); err != nil {
		logger.Log.Error(err.Error())
		return false
	}
	return true
}
