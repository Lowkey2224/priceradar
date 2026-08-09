// internal/db/product_test.go
package db

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const nameAfterChange = "New Bar Name"

func TestProduct_CRUD(t *testing.T) {
	dbConn := setupTest(t)

	newProduct := Product{
		Title: "Test Riegel",
		Urls: StringArray{
			"https://shop-a.com/riegel",
			"https://shop-b.com/riegel",
		},
		TargetPrice: 299,
	}

	err := newProduct.Create(dbConn)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if newProduct.ID == "" {
		t.Errorf("Create() should update p.ID but was empty")
	}

	// --- 2. TEST: GetProduct ---
	fetched, err := GetProduct(dbConn, newProduct.ID)
	if err != nil {
		t.Fatalf("GetProduct() failed: %v", err)
	}

	if fetched.Title != newProduct.Title {
		t.Errorf("GetProduct() Title = %v, want %v", fetched.Title, newProduct.Title)
	}

	if len(fetched.Urls) != 2 {
		t.Errorf("GetProduct() Urls Anzahl = %d, want 2", len(fetched.Urls))
	}

	if fetched.Urls[0] != newProduct.Urls[0] || fetched.Urls[1] != newProduct.Urls[1] {
		t.Errorf("GetProduct() Urls mapped wrongly: %v, want %v", fetched.Urls, newProduct.Urls)
	}

	// --- 3. TEST: Update ---
	fetched.Title = nameAfterChange
	fetched.Urls = append(fetched.Urls, "https://shop-c.com/riegel")
	fetched.TargetPrice = 349

	err = fetched.Update(dbConn)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// --- 4. TEST: Verify Update
	updated, err := GetProduct(dbConn, newProduct.ID)
	if err != nil {
		t.Fatalf("GetProduct() failed after update: %v", err)
	}

	if updated.Title != nameAfterChange {
		t.Errorf("Update() new title wasnt saved. got: %s, want %s", updated.Title, nameAfterChange)
	}

	if len(updated.Urls) != 3 {
		t.Errorf("Update() Urls werent updated. Got len: %d, want 3", len(updated.Urls))
	}
}
