package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"kasir-api/config"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
)

func main() {
	// Load configuration
	cfg, cfgErr := config.LoadConfig()
	if cfgErr != nil {
		panic(cfgErr)
	}

	db, db_err := database.Connect(cfg.DB)
	if db_err != nil {
		panic(db_err)
	}
	defer db.Close()

	// Wire: repository -> service -> handler
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	categoryRepo := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	port := cfg.App.Port
	addr := ":" + strconv.Itoa(port)
	fmt.Printf("Starting server on %s\n", addr)

	// Handle API routes
	http.HandleFunc("/api/products", productHandler.Handle)
	http.HandleFunc("/api/products/", productHandler.Handle)

	http.HandleFunc("/categories", categoryHandler.Handle)
	http.HandleFunc("/categories/", categoryHandler.Handle)

	// Health check
	http.HandleFunc("/health", handleHealth)

	http.HandleFunc("/", handleRoot)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Hello, Ini Backend Program Kasir!"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "Service is running"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
