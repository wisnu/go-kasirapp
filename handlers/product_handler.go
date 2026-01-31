package handlers

import (
	"encoding/json"
	"net/http"

	"kasir-api/models"
	"kasir-api/services"
)

func HandleProducts(w http.ResponseWriter, r *http.Request) {
	// Handle GET, PUT, DELETE /api/products/{id}
	if r.URL.Path != "/api/products" && r.URL.Path != "/api/products/" {
		switch r.Method {
		case http.MethodGet, http.MethodPut, http.MethodDelete:
		default:
			WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		id, err := ParseAndValidateIDFromPath(r.URL.Path, "/api/products/")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			product, err := services.GetProductByID(id)
			if err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, product)
		case http.MethodPut:
			var updated models.Product
			if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
				WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			result, err := services.UpdateProduct(id, updated)
			if err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, result)
		case http.MethodDelete:
			if err := services.DeleteProduct(id); err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
		}
		return
	}

	// Handle GET all products
	if r.Method == http.MethodGet {
		WriteJSON(w, http.StatusOK, services.GetAllProducts())
		return
	}

	// Handle POST to add a new product
	if r.Method == http.MethodPost {
		var newProduct models.Product
		if err := json.NewDecoder(r.Body).Decode(&newProduct); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		created := services.CreateProduct(newProduct)
		WriteJSON(w, http.StatusCreated, created)
		return
	}
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}
