package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Lowkey2224/priceradar/internal/api"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	connStr := os.Getenv("DATABASE_URL")

	dbConn, err := sql.Open("pgx", connStr)
	mux := http.NewServeMux()
	productController := api.ProductController{
		DbConn: dbConn,
	}
	mux.HandleFunc("POST /product", productController.CreateProductHandler)
}
