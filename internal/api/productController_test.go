package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lowkey2224/priceradar/internal/testhelpers"
)

func TestCreateProductHandler_Success(t *testing.T) {
	// 1. Mock/Test-DB aufbauen (z. B. via Interface oder Test-DB-Instanz)
	mockDB := testhelpers.SetupTestDB(t)
	controller := ProductController{DbConn: mockDB}

	// 2. Request-Body vorbereiten
	jsonBody := []byte(`{
		"title": "Test Produkt",
		"urls": ["https://example.com"],
		"targetPrice": 100
	}`)

	// 3. Fake-Request und ResponseRecorder erstellen
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// 4. Handler direkt aufrufen
	controller.CreateProductHandler(rec, req)

	// 5. Assertions durchführen
	if rec.Code != http.StatusOK { // bzw. http.StatusCreated
		t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "Test Produkt") {
		t.Errorf("expected response to contain product title, got %s", rec.Body.String())
	}
}
