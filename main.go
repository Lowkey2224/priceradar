package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Lowkey2224/priceradar/internal/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib" // Registriert den pgx-Treiber für database/sql
	"github.com/joho/godotenv"
)

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("treiber konnte nicht erstellt werden: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migration konnte nicht initialisiert werden: %w", err)
	}

	// 3. Alle ausstehenden .up.sql Dateien ausführen
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Schema ist bereits auf dem neuesten Stand (keine Änderungen).")
			return nil
		}
		return fmt.Errorf("fehler beim Ausführen der Migration: %w", err)
	}

	log.Println("Migrationen erfolgreich angewendet!")
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	connStr := os.Getenv("DATABASE_URL")

	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	if err := runMigrations(dbConn); err != nil {
		log.Fatalf("Migration Error: %v", err)
	}
	product := db.Product{
		Title:       "Energieriegel",
		Urls:        db.StringArray{"https://example.com", "https://amazon.de"},
		TargetPrice: 145,
	}

	err = product.Create(dbConn)
	fmt.Printf("Neuer Eintrag mit ID: %s\n", product.ID)
	fmt.Printf("%v\n", product)

}
