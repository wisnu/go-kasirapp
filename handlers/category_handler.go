package handlers

import (
	"encoding/json"
	"net/http"

	"kasir-api/models"
	"kasir-api/services"
)

func HandleCategories(w http.ResponseWriter, r *http.Request) {
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
			category, err := services.GetCategoryByID(id)
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
			result, err := services.UpdateCategory(id, updated)
			if err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, result)
		case http.MethodDelete:
			if err := services.DeleteCategory(id); err != nil {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"message": "Category deleted"})
		}
		return
	}

	// Handle GET all categories
	if r.Method == http.MethodGet {
		WriteJSON(w, http.StatusOK, services.GetAllCategories())
		return
	}

	// Handle POST to add a new category
	if r.Method == http.MethodPost {
		var newCategory models.Category
		if err := json.NewDecoder(r.Body).Decode(&newCategory); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		created := services.CreateCategory(newCategory)
		WriteJSON(w, http.StatusCreated, created)
		return
	}
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}
