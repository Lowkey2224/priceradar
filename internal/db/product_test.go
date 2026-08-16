// internal/db/product_test.go
package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Lowkey2224/priceradar/internal/testhelpers"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

const nameAfterChange = "New Bar Name"

func TestProduct_CRUD(t *testing.T) {
	ctx := context.Background()
	dbConn := testhelpers.SetupTestDB(t)

	newProduct := Product{
		Title: "Test Riegel",
		Urls: pq.StringArray{
			"https://shop-a.com/riegel",
			"https://shop-b.com/riegel",
		},
		TargetPrice: 299,
	}

	err := newProduct.Create(ctx, dbConn)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if newProduct.ID == "" {
		t.Errorf("Create() should update p.ID but was empty")
	}

	// --- 2. TEST: GetProduct ---
	fetched, err := GetProduct(ctx, dbConn, newProduct.ID)
	if err != nil {
		t.Fatalf("GetProduct() failed: %v", err)
	}

	if fetched.Title != newProduct.Title {
		t.Errorf("GetProduct() Title = %v, want %v", fetched.Title, newProduct.Title)
	}

	if fetched.TargetPrice != newProduct.TargetPrice {
		t.Errorf("GetProduct() TargetPrice = %v, want %v", fetched.TargetPrice, newProduct.TargetPrice)
	}

	if len(fetched.Urls) != 2 {
		t.Fatalf("GetProduct() Urls Anzahl = %d, want 2", len(fetched.Urls))
	}

	if fetched.Urls[0] != newProduct.Urls[0] || fetched.Urls[1] != newProduct.Urls[1] {
		t.Errorf("GetProduct() Urls mapped wrongly: %v, want %v", fetched.Urls, newProduct.Urls)
	}

	// --- 3. TEST: Update ---
	fetched.Title = nameAfterChange
	fetched.Urls = append(fetched.Urls, "https://shop-c.com/riegel")
	fetched.TargetPrice = 349

	err = fetched.Update(ctx, dbConn)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// --- 4. TEST: Verify Update
	updated, err := GetProduct(ctx, dbConn, newProduct.ID)
	if err != nil {
		t.Fatalf("GetProduct() failed after update: %v", err)
	}

	if updated.Title != nameAfterChange {
		t.Errorf("Update() new title wasnt saved. got: %s, want %s", updated.Title, nameAfterChange)
	}

	if updated.TargetPrice != 349 {
		t.Errorf("Update() new targetPrice wasnt saved. got: %d, want 349", updated.TargetPrice)
	}

	if len(updated.Urls) != 3 {
		t.Fatalf("Update() Urls werent updated. Got len: %d, want 3", len(updated.Urls))
	}

	if updated.Urls[0] != newProduct.Urls[0] ||
		updated.Urls[1] != newProduct.Urls[1] ||
		updated.Urls[2] != "https://shop-c.com/riegel" {
		t.Errorf("Update() URLs were not persisted: got %v", updated.Urls)
	}

	err = updated.Delete(ctx, dbConn)

	if err != nil {
		t.Fatalf("Delete() failed with error %v", err)
	}
	_, err = GetProduct(ctx, dbConn, newProduct.ID)

	if err != sql.ErrNoRows {
		t.Fatalf("wrong error got %v expected %v", err, sql.ErrNoRows)
	}

}
