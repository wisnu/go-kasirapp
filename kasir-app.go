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

// Mock data
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
	http.HandleFunc("/api/products/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/api/products/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		for _, product := range products {
			if product.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(product)
				return
			}
		}
		http.Error(w, "Product not found", http.StatusNotFound)
	})

	// Handle GET /api/products to get list all products and POST /api/products to add a new product
	// GET /api/products - List all products
	// POST /api/products - Add a new product
	http.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(products)
		} else if r.Method == http.MethodPost {
			var newProduct Product
			err := json.NewDecoder(r.Body).Decode(&newProduct)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			newProduct.ID = len(products) + 1
			products = append(products, newProduct)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newProduct)
		}

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, Ini Backend Program Kasir!"})
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
