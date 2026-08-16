package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

const tableName = "products"
const (
	colID          = "id"
	colTitle       = "title"
	colUrls        = "urls"
	colTargetPrice = "target_price"
	colCreatedAt   = "created_at"
	colUpdatedAt   = "updated_at"
)

const insertCols = colTitle + ", " + colUrls + ", " + colTargetPrice
const selectCols = colID + ", " + insertCols + ", " + colCreatedAt + ", " + colUpdatedAt

type Product struct {
	ID          string
	Title       string
	Urls        pq.StringArray
	TargetPrice uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Product) Create(ctx context.Context, db *sql.DB) error {

	query := `INSERT INTO ` + tableName + ` (` + insertCols + `) VALUES ($1, $2, $3) RETURNING ` + colID + ", " + colCreatedAt + ", " + colUpdatedAt

	err := db.QueryRowContext(
		ctx,
		query,
		p.Title,
		p.Urls,
		p.TargetPrice,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	return err
}

func (p Product) Delete(ctx context.Context, db *sql.DB) error {
	query := `DELETE FROM ` + tableName + ` where ` + colID + ` = $1`
	result, err := db.ExecContext(ctx, query, p.ID)
	if err != nil {
		return fmt.Errorf("error executing delete: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching affectedRows: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (p *Product) Update(ctx context.Context, db *sql.DB) error {
	updatedAt := time.Now()

	query := `
		UPDATE ` + tableName + ` 
		SET ` +
		colTitle + ` = $1, ` +
		colUrls + ` = $2, ` +
		colTargetPrice + ` = $3, ` +
		colUpdatedAt + ` = $4 
		WHERE ` + colID + ` = $5`

	result, err := db.ExecContext(
		ctx,
		query,
		p.Title,
		p.Urls,
		p.TargetPrice,
		updatedAt,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("error executing update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching affectedRows: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	p.UpdatedAt = updatedAt
	return nil
}

func GetProduct(ctx context.Context, db *sql.DB, id string) (Product, error) {
	var p Product

	selectQuery := `SELECT ` + selectCols + ` FROM ` + tableName + ` WHERE id = $1`
	err := db.QueryRowContext(ctx, selectQuery, id).Scan(
		&p.ID,
		&p.Title,
		&p.Urls,
		&p.TargetPrice,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	return p, err
}
