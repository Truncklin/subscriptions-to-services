package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPool(storagePath string) (*pgxpool.Pool, error) {
	dbConf, err := pgxpool.ParseConfig(storagePath)
	if err != nil {
		return nil, err
	}

	dbConf.MaxConns = 10

	var pool *pgxpool.Pool
	var errPing error

	for i := 1; i <= 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err = pgxpool.NewWithConfig(ctx, dbConf)
		cancel()
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}

		ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
		errPing = pool.Ping(ctxPing)
		cancelPing()

		if errPing == nil {
			return pool, nil
		}

		pool.Close()
		time.Sleep(time.Duration(i) * time.Second)
	}

	return nil, errPing
}

func RunMigrations(storagePath string) error {

	migrationsPath := "/migrations"
	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
