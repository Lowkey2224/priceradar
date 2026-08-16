package testhelpers

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

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connStr := os.Getenv("TEST_DB_URL")
	if connStr == "" {
		connStr = "postgres://price-alert:price-alert@localhost:5432/price-alert?sslmode=disable"
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Connection to DB Failed: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatalf("Migration driver error: %v", err)
	}

	mig, err := migrate.NewWithDatabaseInstance("file://../db/migrations", "postgres", driver)
	if err != nil {
		t.Fatalf("Migration Error: %v", err)
	}

	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("Error executing migration: %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.Exec("TRUNCATE TABLE products CASCADE"); err != nil {
			t.Errorf("cleanup: truncate products: %v", err)
		}
		_ = db.Close()
	})

	return db
}
