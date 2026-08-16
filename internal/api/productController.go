package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/Lowkey2224/priceradar/internal/api/dto"
	"github.com/Lowkey2224/priceradar/internal/db"
)

type ProductController struct {
	DbConn *sql.DB
}

func (c ProductController) CreateProductHandler(w http.ResponseWriter, r *http.Request) {

	request := dto.CreateProductRequest{}

	if err := request.New(r.Body); err != nil {
		http.Error(w, "Body is not valid JSON", http.StatusBadRequest)
		return
	}
	if err := request.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product := db.Product{
		Title:       request.Title,
		Urls:        request.Urls,
		TargetPrice: uint64(request.TargetPrice),
	}

	if err := product.Create(context.Background(), c.DbConn); err != nil {
		log.Printf("Error writing product into DB with error:\n%v\n", err)
		http.Error(w, "Error writing Product into DB", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, dto.CreateProductResponseFromProduct(product).ToJson())
}
