package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"kasir-api/config"
	"kasir-api/database"
	"kasir-api/models"
)

// Mock data for products
var products = []models.Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Stock: 10},
	{ID: 2, Name: "Smartphone", Price: 499.99, Stock: 25},
	{ID: 3, Name: "Tablet", Price: 299.99, Stock: 15},
	{ID: 4, Name: "Headphones", Price: 99.99, Stock: 60},
}

// Mock data for categories
var categories = []models.Category{
	{ID: 1, Name: "Electronics", Description: "Electronic devices and gadgets"},
	{ID: 2, Name: "Accessories", Description: "Related accessories and add-ons"},
}

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

	port := cfg.App.Port
	addr := ":" + strconv.Itoa(port)
	fmt.Printf("Starting server on %s\n", addr)

	// Handle API routes

	// Handle GET /api/products to get list all products and POST /api/products to add a new product
	// GET /api/products - List all products
	// POST /api/products - Add a new product
	http.HandleFunc("/api/products", handleProducts)
	http.HandleFunc("/api/products/", handleProducts)

	// Handle CRUD /categories and /categories/{id}
	http.HandleFunc("/categories", handleCategories)
	http.HandleFunc("/categories/", handleCategories)

	// Health check
	http.HandleFunc("/health", handleHealth)

	http.HandleFunc("/", handleRoot)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	// Handle GET , PUT, DELETE /api/products/{id}
	if r.URL.Path != "/api/products" && r.URL.Path != "/api/products/" {
		switch r.Method {
		case http.MethodGet, http.MethodPut, http.MethodDelete:
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		id, err := parseAndValidateIDFromPath(r.URL.Path, "/api/products/")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			for _, product := range products {
				if product.ID == id {
					writeJSON(w, http.StatusOK, product)
					return
				}
			}
			writeError(w, http.StatusNotFound, "Product not found")
		case http.MethodPut:
			var updated models.Product
			err := json.NewDecoder(r.Body).Decode(&updated)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			for i, product := range products {
				if product.ID == id {
					updated.ID = product.ID
					products[i] = updated
					writeJSON(w, http.StatusOK, updated)
					return
				}
			}
			writeError(w, http.StatusNotFound, "Product not found")
		case http.MethodDelete:
			for i, product := range products {
				if product.ID == id {
					products = append(products[:i], products[i+1:]...)
					writeJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
					return
				}
			}
			writeError(w, http.StatusNotFound, "Product not found")
		}
		return
	}

	// Handle GET all products
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, products)
		return
	}

	// Handle POST to add a new product
	if r.Method == http.MethodPost {
		var newProduct models.Product
		err := json.NewDecoder(r.Body).Decode(&newProduct)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		newProduct.ID = nextProductID()
		products = append(products, newProduct)
		writeJSON(w, http.StatusCreated, newProduct)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
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

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func handleCategories(w http.ResponseWriter, r *http.Request) {

	// Handle GET , PUT, DELETE /categories/{id}
	if r.URL.Path != "/categories" && r.URL.Path != "/categories/" {
		if r.Method != http.MethodGet && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		id, err := parseAndValidateIDFromPath(r.URL.Path, "/categories/")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}

		for i, category := range categories {
			if category.ID == id {
				switch r.Method {
				case http.MethodGet:
					writeJSON(w, http.StatusOK, category)
					return
				case http.MethodPut:
					var updated models.Category
					err := json.NewDecoder(r.Body).Decode(&updated)
					if err != nil {
						writeError(w, http.StatusBadRequest, err.Error())
						return
					}
					updated.ID = category.ID
					categories[i] = updated
					writeJSON(w, http.StatusOK, updated)
					return
				case http.MethodDelete:
					categories = append(categories[:i], categories[i+1:]...)
					writeJSON(w, http.StatusOK, map[string]string{"message": "Category deleted"})
					return
				}
			}
		}
		writeError(w, http.StatusNotFound, "Category not found")
		return
	}

	// Handle GET all categories
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, categories)
		return
	}

	// Handle POST to add a new category
	if r.Method == http.MethodPost {
		var newCategory models.Category
		err := json.NewDecoder(r.Body).Decode(&newCategory)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		newCategory.ID = nextCategoryID()
		categories = append(categories, newCategory)
		writeJSON(w, http.StatusCreated, newCategory)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func nextCategoryID() int {
	maxID := 0
	for _, category := range categories {
		if category.ID > maxID {
			maxID = category.ID
		}
	}
	return maxID + 1
}

func nextProductID() int {
	maxID := 0
	for _, product := range products {
		if product.ID > maxID {
			maxID = product.ID
		}
	}
	return maxID + 1
}

func parseAndValidateIDFromPath(path, prefix string) (int, error) {
	// Extract ID from path
	idPart := strings.TrimPrefix(path, prefix)

	// Validate ID part, must be a number
	if idPart == "" || strings.Contains(idPart, "/") {
		return 0, fmt.Errorf("invalid path")
	}
	return strconv.Atoi(idPart)
}
