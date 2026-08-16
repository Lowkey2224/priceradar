package dto

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lowkey2224/priceradar/internal/db"
)

func TestDesrialization(t *testing.T) {
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-1 * time.Hour)
	product := db.Product{
		ID:          "some-uuid",
		Title:       "ExpectedTitle",
		Urls:        []string{"https://foo.bar", "http://example.com"},
		TargetPrice: 123,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	expectedJson := fmt.Sprintf(
		`{"id":"some-uuid","title":"ExpectedTitle","urls":["https://foo.bar","http://example.com"],"targetPrice":123,"createdAt":"%s","updatedAt":"%s"}`,
		createdAt.Format(time.RFC3339Nano),
		updatedAt.Format(time.RFC3339Nano),
	)
	response := CreateProductResponseFromProduct(product)
	actualJson := response.ToJson()
	if actualJson != expectedJson {
		t.Fatalf("ToJson() expected \n%s\n, got \n%s", expectedJson, actualJson)
	}
}
