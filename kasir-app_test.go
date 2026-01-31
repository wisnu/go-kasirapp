package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"kasir-api/handlers"
	"kasir-api/models"
	"kasir-api/repositories"
	"kasir-api/services"
)

// setupProductHandler creates a ProductHandler wired to a sqlmock DB.
func setupProductHandler(t *testing.T) (*handlers.ProductHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := repositories.NewProductRepository(db)
	svc := services.NewProductService(repo)
	return handlers.NewProductHandler(svc), mock
}

// setupCategoryHandler creates a CategoryHandler wired to a sqlmock DB.
func setupCategoryHandler(t *testing.T) (*handlers.CategoryHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := repositories.NewCategoryRepository(db)
	svc := services.NewCategoryService(repo)
	return handlers.NewCategoryHandler(svc), mock
}

func doRequest(t *testing.T, method, path string, body interface{}, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestRootAndHealth(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/", nil, handleRoot)
	if rec.Code != http.StatusOK {
		t.Fatalf("root status = %d, want %d", rec.Code, http.StatusOK)
	}

	rec = doRequest(t, http.MethodGet, "/health", nil, handleHealth)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestProductsListAndCreate(t *testing.T) {
	handler, mock := setupProductHandler(t)

	// --- GET /api/products ---
	rows := sqlmock.NewRows([]string{"id", "name", "price", "stock"}).
		AddRow(1, "Laptop", 999.99, 10).
		AddRow(2, "Smartphone", 499.99, 25).
		AddRow(3, "Tablet", 299.99, 15).
		AddRow(4, "Headphones", 99.99, 60)
	mock.ExpectQuery("SELECT id, name, price, stock FROM products").WillReturnRows(rows)

	rec := doRequest(t, http.MethodGet, "/api/products", nil, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("list products status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []models.Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("list products len = %d, want 4", len(got))
	}

	// --- POST /api/products ---
	mock.ExpectQuery("INSERT INTO products").
		WithArgs("Mouse", 25.5, 50).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	newProduct := models.Product{Name: "Mouse", Price: 25.5, Stock: 50}
	rec = doRequest(t, http.MethodPost, "/api/products", newProduct, handler.Handle)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var created models.Product
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID != 5 {
		t.Fatalf("created product id = %d, want 5", created.ID)
	}
	if created.Name != newProduct.Name {
		t.Fatalf("created product name = %q, want %q", created.Name, newProduct.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProductsGetUpdateDelete(t *testing.T) {
	handler, mock := setupProductHandler(t)

	// --- GET /api/products/1 ---
	mock.ExpectQuery("SELECT id, name, price, stock FROM products WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "stock"}).AddRow(1, "Laptop", 999.99, 10))

	rec := doRequest(t, http.MethodGet, "/api/products/1", nil, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("get product status = %d, want %d", rec.Code, http.StatusOK)
	}

	// --- PUT /api/products/1 ---
	mock.ExpectExec("UPDATE products SET").
		WithArgs("Laptop Pro", 1299.99, 7, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	update := models.Product{Name: "Laptop Pro", Price: 1299.99, Stock: 7}
	rec = doRequest(t, http.MethodPut, "/api/products/1", update, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("update product status = %d, want %d", rec.Code, http.StatusOK)
	}
	var updated models.Product
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.ID != 1 || updated.Name != update.Name {
		t.Fatalf("updated product = %+v, want id=1 name=%q", updated, update.Name)
	}

	// --- DELETE /api/products/1 ---
	mock.ExpectExec("DELETE FROM products WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec = doRequest(t, http.MethodDelete, "/api/products/1", nil, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete product status = %d, want %d", rec.Code, http.StatusOK)
	}

	// --- GET /api/products/1 after delete → 404 ---
	mock.ExpectQuery("SELECT id, name, price, stock FROM products WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "stock"})) // empty result
	rec = doRequest(t, http.MethodGet, "/api/products/1", nil, handler.Handle)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted product status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoriesListAndCreate(t *testing.T) {
	handler, mock := setupCategoryHandler(t)

	// --- GET /categories ---
	rows := sqlmock.NewRows([]string{"id", "name", "description"}).
		AddRow(1, "Electronics", "Electronic devices and gadgets").
		AddRow(2, "Accessories", "Related accessories and add-ons")
	mock.ExpectQuery("SELECT id, name, description FROM categories").WillReturnRows(rows)

	rec := doRequest(t, http.MethodGet, "/categories", nil, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("list categories status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []models.Category
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("list categories len = %d, want 2", len(got))
	}

	// --- POST /categories ---
	mock.ExpectQuery("INSERT INTO categories").
		WithArgs("Office", "Office equipment").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	newCategory := models.Category{Name: "Office", Description: "Office equipment"}
	rec = doRequest(t, http.MethodPost, "/categories", newCategory, handler.Handle)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create category status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var created models.Category
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID != 3 {
		t.Fatalf("created category id = %d, want 3", created.ID)
	}
	if created.Name != newCategory.Name {
		t.Fatalf("created category name = %q, want %q", created.Name, newCategory.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCategoriesGetUpdateDelete(t *testing.T) {
	handler, mock := setupCategoryHandler(t)

	// --- GET /categories/1 ---
	mock.ExpectQuery("SELECT id, name, description FROM categories WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "Electronics", "Electronic devices and gadgets"))

	rec := doRequest(t, http.MethodGet, "/categories/1", nil, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("get category status = %d, want %d", rec.Code, http.StatusOK)
	}

	// --- PUT /categories/1 ---
	mock.ExpectExec("UPDATE categories SET").
		WithArgs("Electronics+", "Updated description", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	update := models.Category{Name: "Electronics+", Description: "Updated description"}
	rec = doRequest(t, http.MethodPut, "/categories/1", update, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("update category status = %d, want %d", rec.Code, http.StatusOK)
	}
	var updated models.Category
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.ID != 1 || updated.Name != update.Name {
		t.Fatalf("updated category = %+v, want id=1 name=%q", updated, update.Name)
	}

	// --- DELETE /categories/1 ---
	mock.ExpectExec("DELETE FROM categories WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec = doRequest(t, http.MethodDelete, "/categories/1", nil, handler.Handle)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete category status = %d, want %d", rec.Code, http.StatusOK)
	}

	// --- GET /categories/1 after delete → 404 ---
	mock.ExpectQuery("SELECT id, name, description FROM categories WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"})) // empty result
	rec = doRequest(t, http.MethodGet, "/categories/1", nil, handler.Handle)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted category status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
