package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Lowkey2224/priceradar/internal/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib" // Registriert den pgx-Treiber für database/sql
)

func runMigrations(db *sql.DB) error {
	// 1. Postgres-Treiber für golang-migrate initialisieren
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("treiber konnte nicht erstellt werden: %w", err)
	}

	// 2. Migration-Instanz erstellen (liest aus dem Ordner "migrations")
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migration konnte nicht initialisiert werden: %w", err)
	}

	// 3. Alle ausstehenden .up.sql Dateien ausführen
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Schema ist bereits auf dem neuesten Stand (keine Änderungen).")
			return nil
		}
		return fmt.Errorf("fehler beim Ausführen der Migration: %w", err)
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
	connStr := "postgres://price-alert:price-alert@localhost:5432/price-alert?sslmode=disable"

	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

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
	pNew, err := db.GetProduct(dbConn, product.ID)
	if err != nil {
		log.Fatalf("Fetch error: %v\n", err)
	}

	fmt.Printf("Product fetched %v\n", pNew)
	pNew.Urls = append(pNew.Urls, "https://www.amazon.de")
	err = pNew.Update(dbConn)
	if err != nil {
		log.Fatalf("Update error: %v\n", err)
	}

	fmt.Printf("in Memory fetched %v\n", pNew)
	pDoubleNew, err := db.GetProduct(dbConn, product.ID)

	if err != nil {
		log.Fatalf("Fetch error: %v\n", err)
	}
	fmt.Printf("newly fetched %v\n", pDoubleNew)

}
