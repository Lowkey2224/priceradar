// internal/db/main_test.go
package db

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	connStr := os.Getenv("TEST_DB_URL")
	if connStr == "" {
		// Fallback für lokales Testen
		connStr = "postgres://price-alert:price-alert@localhost:5432/price-alert?sslmode=disable"
	}

	var err error
	testDB, err = sql.Open("pgx", connStr)
	if err != nil {
		panic("Connection to DB Failed: " + err.Error())
	}

	driver, err := postgres.WithInstance(testDB, &postgres.Config{})
	if err != nil {
		panic("Migration driver error: " + err.Error())
	}

	mig, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		panic("Migration Error: " + err.Error())
	}

	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic("Error executing migration: " + err.Error())
	}

	code := m.Run()

	testDB.Close()

	os.Exit(code)
}

func setupTest(t *testing.T) *sql.DB {
	t.Helper()

	t.Cleanup(func() {
		_, _ = testDB.Exec("TRUNCATE TABLE products CASCADE")
	})

	return testDB
}
