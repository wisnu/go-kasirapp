package services

import (
	"errors"
	"kasir-api/models"
)

// Categories is the in-memory data store for categories.
var Categories = []models.Category{
	{ID: 1, Name: "Electronics", Description: "Electronic devices and gadgets"},
	{ID: 2, Name: "Accessories", Description: "Related accessories and add-ons"},
}

func GetAllCategories() []models.Category {
	return Categories
}

func GetCategoryByID(id int) (models.Category, error) {
	for _, category := range Categories {
		if category.ID == id {
			return category, nil
		}
	}
	return models.Category{}, errors.New("Category not found")
}

func CreateCategory(category models.Category) models.Category {
	category.ID = nextCategoryID()
	Categories = append(Categories, category)
	return category
}

func UpdateCategory(id int, updated models.Category) (models.Category, error) {
	for i, category := range Categories {
		if category.ID == id {
			updated.ID = category.ID
			Categories[i] = updated
			return updated, nil
		}
	}
	return models.Category{}, errors.New("Category not found")
}

func DeleteCategory(id int) error {
	for i, category := range Categories {
		if category.ID == id {
			Categories = append(Categories[:i], Categories[i+1:]...)
			return nil
		}
	}
	return errors.New("Category not found")
}

func nextCategoryID() int {
	maxID := 0
	for _, category := range Categories {
		if category.ID > maxID {
			maxID = category.ID
		}
	}
	return maxID + 1
}
