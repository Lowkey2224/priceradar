package dto

import (
	"encoding/json"
	"time"

	"github.com/Lowkey2224/priceradar/internal/db"
)

type ProductResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Urls        []string  `json:"urls"`
	TargetPrice uint64    `json:"targetPrice"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func CreateProductResponseFromProduct(product db.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Title:       product.Title,
		Urls:        product.Urls,
		TargetPrice: product.TargetPrice,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func (p ProductResponse) ToJson() string {
	b, _ := json.Marshal(p)
	return string(b)
}
