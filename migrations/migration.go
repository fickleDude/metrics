package migrations

import (
	"embed"
	"sync"

	"github.com/fickleDude/metrics.git/internal/config/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var files embed.FS

type Migrator struct {
	srcDriver source.Driver
	dbDriver  database.Driver
}

var (
	instance *Migrator
	once     sync.Once
)

func GetMigrator() *Migrator {
	once.Do(func() {
		//init db driver
		db, err := postgres.WithInstance(db.GetDBConnection(), &postgres.Config{})
		if err != nil {
			return
		}
		//init source driver
		src, err := iofs.New(files, ".")
		if err != nil {
			return
		}
		instance = &Migrator{srcDriver: src, dbDriver: db}
	})
	return instance
}

func (m *Migrator) MigrateUp() error {

	//create migration instance
	i, err := migrate.NewWithInstance("iofs", m.srcDriver, "postgres", m.dbDriver)
	if err != nil {
		return err
	}
	//migrate
	err = i.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
func (m *Migrator) MigrateDown() error {
	//create migration instance
	i, err := migrate.NewWithInstance("iofs", m.srcDriver, "postgres", m.dbDriver)
	if err != nil {
		return err
	}
	//migrate
	err = i.Down()
	if err != nil {
		return err
	}
	return nil
}
