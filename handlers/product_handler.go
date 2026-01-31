package handlers

import (
	"encoding/json"
	"net/http"

	"kasir-api/models"
	"kasir-api/services"
)

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
			product, err := h.service.GetByID(id)
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
			updated.ID = id
			if err := h.service.Update(&updated); err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, updated)
		case http.MethodDelete:
			if err := h.service.Delete(id); err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
		}
		return
	}

	// Handle GET all products
	if r.Method == http.MethodGet {
		products, err := h.service.GetAll()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, products)
		return
	}

	// Handle POST to add a new product
	if r.Method == http.MethodPost {
		var newProduct models.Product
		if err := json.NewDecoder(r.Body).Decode(&newProduct); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := h.service.Create(&newProduct); err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusCreated, newProduct)
		return
	}
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}
