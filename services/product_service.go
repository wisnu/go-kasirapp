package services

import (
	"errors"
	"kasir-api/models"
)

// Products is the in-memory data store for products.
var Products = []models.Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Stock: 10},
	{ID: 2, Name: "Smartphone", Price: 499.99, Stock: 25},
	{ID: 3, Name: "Tablet", Price: 299.99, Stock: 15},
	{ID: 4, Name: "Headphones", Price: 99.99, Stock: 60},
}

func GetAllProducts() []models.Product {
	return Products
}

func GetProductByID(id int) (models.Product, error) {
	for _, product := range Products {
		if product.ID == id {
			return product, nil
		}
	}
	return models.Product{}, errors.New("Product not found")
}

func CreateProduct(product models.Product) models.Product {
	product.ID = nextProductID()
	Products = append(Products, product)
	return product
}

func UpdateProduct(id int, updated models.Product) (models.Product, error) {
	for i, product := range Products {
		if product.ID == id {
			updated.ID = product.ID
			Products[i] = updated
			return updated, nil
		}
	}
	return models.Product{}, errors.New("Product not found")
}

func DeleteProduct(id int) error {
	for i, product := range Products {
		if product.ID == id {
			Products = append(Products[:i], Products[i+1:]...)
			return nil
		}
	}
	return errors.New("Product not found")
}

func nextProductID() int {
	maxID := 0
	for _, product := range Products {
		if product.ID > maxID {
			maxID = product.ID
		}
	}
	return maxID + 1
}
