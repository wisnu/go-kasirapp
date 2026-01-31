package handlers

import (
	"encoding/json"
	"net/http"

	"kasir-api/models"
	"kasir-api/services"
)

type CategoryHandler struct {
	service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Handle GET, PUT, DELETE /categories/{id}
	if r.URL.Path != "/categories" && r.URL.Path != "/categories/" {
		if r.Method != http.MethodGet && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		id, err := ParseAndValidateIDFromPath(r.URL.Path, "/categories/")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			category, err := h.service.GetByID(id)
			if err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, category)
		case http.MethodPut:
			var updated models.Category
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
			WriteJSON(w, http.StatusOK, map[string]string{"message": "Category deleted"})
		}
		return
	}

	// Handle GET all categories
	if r.Method == http.MethodGet {
		categories, err := h.service.GetAll()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, categories)
		return
	}

	// Handle POST to add a new category
	if r.Method == http.MethodPost {
		var newCategory models.Category
		if err := json.NewDecoder(r.Body).Decode(&newCategory); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := h.service.Create(&newCategory); err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteJSON(w, http.StatusCreated, newCategory)
		return
	}
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}
