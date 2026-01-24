package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// Represents a product in the inventory
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// Mock data for products
var products = []Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Stock: 10},
	{ID: 2, Name: "Smartphone", Price: 499.99, Stock: 25},
	{ID: 3, Name: "Tablet", Price: 299.99, Stock: 15},
	{ID: 4, Name: "Headphones", Price: 99.99, Stock: 60},
}

func main() {
	fmt.Println("Starting server on :8080")

	// Handle API routes

	// Handle GET /api/products/{id} to get product by ID
	// GET /api/products/{id} - Get product by ID
	http.HandleFunc("/api/products/", handleGetProductByID)

	// Handle GET /api/products to get list all products and POST /api/products to add a new product
	// GET /api/products - List all products
	// POST /api/products - Add a new product
	http.HandleFunc("/api/products", handleProducts)

	http.HandleFunc("/", handleRoot)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/products/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	for _, product := range products {
		if product.ID == id {
			writeJSON(w, http.StatusOK, product)
			return
		}
	}
	writeError(w, http.StatusNotFound, "Product not found")
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, products)
		return
	}
	if r.Method == http.MethodPost {
		var newProduct Product
		err := json.NewDecoder(r.Body).Decode(&newProduct)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		newProduct.ID = len(products) + 1
		products = append(products, newProduct)
		writeJSON(w, http.StatusCreated, newProduct)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Hello, Ini Backend Program Kasir!"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
