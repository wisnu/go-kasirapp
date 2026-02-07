package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"kasir-api/handlers"
	"kasir-api/models"
	"kasir-api/repositories"
	"kasir-api/services"
)

// baseURL is non-empty when running against a deployed service.
// Set via: BASE_URL=http://host:port go test ./...
var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("BASE_URL")
	os.Exit(m.Run())
}

func isIntegration() bool { return baseURL != "" }

// ---------------------------------------------------------------------------
// Unit-test helpers (sqlmock wiring) — skipped in integration mode
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// doRequest — dispatches to a live server or an in-process handler.
//
//	In integration mode the handler argument is ignored; the request goes to
//	baseURL + path via http.Client.
//	In unit mode it works exactly as before (httptest).
//
// ---------------------------------------------------------------------------
func doRequest(t *testing.T, method, path string, body interface{}, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	if isIntegration() {
		// --- live HTTP call ---
		url := baseURL + path
		req, err := http.NewRequest(method, url, &buf)
		if err != nil {
			t.Fatalf("http.NewRequest: %v", err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("http.Do %s %s: %v", method, url, err)
		}
		defer resp.Body.Close()

		// Pack into a ResponseRecorder so the rest of the test is unchanged.
		rec := httptest.NewRecorder()
		rec.Code = resp.StatusCode
		respBody, _ := io.ReadAll(resp.Body)
		rec.Body.Write(respBody)
		return rec
	}

	// --- in-process (unit) ---
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

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
	var handler http.HandlerFunc
	if !isIntegration() {
		h, mock := setupProductHandler(t)
		handler = h.Handle

		// --- GET /api/products ---
		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "category_name"}).
			AddRow(1, "Laptop", 999.99, 10, "Electronics").
			AddRow(2, "Smartphone", 499.99, 25, "Electronics").
			AddRow(3, "Tablet", 299.99, 15, "Electronics").
			AddRow(4, "Headphones", 99.99, 60, "Accessories")
		mock.ExpectQuery("SELECT p.id, p.name, p.price, p.stock, c.name").WillReturnRows(rows)

		defer func() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		}()

		// --- POST /api/products ---
		mock.ExpectQuery("INSERT INTO products").
			WithArgs("Mouse", 25.5, 50, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	}

	// GET all
	rec := doRequest(t, http.MethodGet, "/api/products", nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("list products status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []models.Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if !isIntegration() && len(got) != 4 {
		t.Fatalf("list products len = %d, want 4", len(got))
	}
	if len(got) == 0 {
		t.Fatalf("list products returned empty")
	}

	// POST
	newProduct := models.Product{Name: "Mouse", Price: 25.5, Stock: 50, CategoryID: 1}
	rec = doRequest(t, http.MethodPost, "/api/products", newProduct, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var created models.Product
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if !isIntegration() && created.ID != 5 {
		t.Fatalf("created product id = %d, want 5", created.ID)
	}
	if created.ID == 0 {
		t.Fatalf("created product has no id")
	}
	if created.Name != newProduct.Name {
		t.Fatalf("created product name = %q, want %q", created.Name, newProduct.Name)
	}

	// Clean up in integration mode: delete the product we just created.
	if isIntegration() {
		t.Logf("integration cleanup: DELETE /api/products/%d", created.ID)
		doRequest(t, http.MethodDelete, "/api/products/"+itoa(created.ID), nil, nil)
	}
}

func TestProductsGetUpdateDelete(t *testing.T) {
	var handler http.HandlerFunc
	var targetID int

	if isIntegration() {
		// Create a product to operate on.
		newProduct := models.Product{Name: "TestItem", Price: 10.0, Stock: 5, CategoryID: 1}
		rec := doRequest(t, http.MethodPost, "/api/products", newProduct, nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("setup create status = %d, want %d", rec.Code, http.StatusCreated)
		}
		var p models.Product
		if err := json.NewDecoder(rec.Body).Decode(&p); err != nil {
			t.Fatalf("decode setup product: %v", err)
		}
		targetID = p.ID
		t.Cleanup(func() {
			// best-effort cleanup
			doRequest(t, http.MethodDelete, "/api/products/"+itoa(targetID), nil, nil)
		})
	} else {
		h, mock := setupProductHandler(t)
		handler = h.Handle
		targetID = 1

		mock.ExpectQuery("SELECT p.id, p.name, p.price, p.stock, c.name").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "stock", "category_name"}).AddRow(1, "Laptop", 999.99, 10, "Electronics"))

		mock.ExpectExec("UPDATE products SET").
			WithArgs("Laptop Pro", 1299.99, 7, 2, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("DELETE FROM products WHERE id").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery("SELECT p.id, p.name, p.price, p.stock, c.name").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "stock", "category_name"}))

		defer func() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		}()
	}

	idStr := itoa(targetID)

	// GET by ID
	rec := doRequest(t, http.MethodGet, "/api/products/"+idStr, nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("get product status = %d, want %d", rec.Code, http.StatusOK)
	}

	// PUT
	update := models.Product{Name: "Laptop Pro", Price: 1299.99, Stock: 7, CategoryID: 2}
	rec = doRequest(t, http.MethodPut, "/api/products/"+idStr, update, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("update product status = %d, want %d", rec.Code, http.StatusOK)
	}
	var updated models.Product
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.ID != targetID || updated.Name != update.Name {
		t.Fatalf("updated product = %+v, want id=%d name=%q", updated, targetID, update.Name)
	}

	// DELETE
	rec = doRequest(t, http.MethodDelete, "/api/products/"+idStr, nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete product status = %d, want %d", rec.Code, http.StatusOK)
	}

	// GET after delete → 404
	rec = doRequest(t, http.MethodGet, "/api/products/"+idStr, nil, handler)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted product status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCategoriesListAndCreate(t *testing.T) {
	var handler http.HandlerFunc
	if !isIntegration() {
		h, mock := setupCategoryHandler(t)
		handler = h.Handle

		rows := sqlmock.NewRows([]string{"id", "name", "description"}).
			AddRow(1, "Electronics", "Electronic devices and gadgets").
			AddRow(2, "Accessories", "Related accessories and add-ons")
		mock.ExpectQuery("SELECT id, name, description FROM categories").WillReturnRows(rows)

		mock.ExpectQuery("INSERT INTO categories").
			WithArgs("Office", "Office equipment").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

		defer func() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		}()
	}

	// GET all
	rec := doRequest(t, http.MethodGet, "/categories", nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("list categories status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []models.Category
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if !isIntegration() && len(got) != 2 {
		t.Fatalf("list categories len = %d, want 2", len(got))
	}
	if len(got) == 0 {
		t.Fatalf("list categories returned empty")
	}

	// POST
	newCategory := models.Category{Name: "Office", Description: "Office equipment"}
	rec = doRequest(t, http.MethodPost, "/categories", newCategory, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create category status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var created models.Category
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if !isIntegration() && created.ID != 3 {
		t.Fatalf("created category id = %d, want 3", created.ID)
	}
	if created.ID == 0 {
		t.Fatalf("created category has no id")
	}
	if created.Name != newCategory.Name {
		t.Fatalf("created category name = %q, want %q", created.Name, newCategory.Name)
	}

	// Clean up in integration mode.
	if isIntegration() {
		t.Logf("integration cleanup: DELETE /categories/%d", created.ID)
		doRequest(t, http.MethodDelete, "/categories/"+itoa(created.ID), nil, nil)
	}
}

func TestCategoriesGetUpdateDelete(t *testing.T) {
	var handler http.HandlerFunc
	var targetID int

	if isIntegration() {
		// Create a category to operate on.
		newCat := models.Category{Name: "TempCat", Description: "Temporary for test"}
		rec := doRequest(t, http.MethodPost, "/categories", newCat, nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("setup create category status = %d, want %d", rec.Code, http.StatusCreated)
		}
		var c models.Category
		if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
			t.Fatalf("decode setup category: %v", err)
		}
		targetID = c.ID
		// No t.Cleanup delete here — the test itself deletes targetID.
	} else {
		h, mock := setupCategoryHandler(t)
		handler = h.Handle
		targetID = 1

		mock.ExpectQuery("SELECT id, name, description FROM categories WHERE id").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "Electronics", "Electronic devices and gadgets"))

		mock.ExpectExec("UPDATE categories SET").
			WithArgs("Electronics+", "Updated description", 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("DELETE FROM categories WHERE id").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery("SELECT id, name, description FROM categories WHERE id").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}))

		defer func() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		}()
	}

	idStr := itoa(targetID)

	// GET by ID
	rec := doRequest(t, http.MethodGet, "/categories/"+idStr, nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("get category status = %d, want %d", rec.Code, http.StatusOK)
	}

	// PUT
	update := models.Category{Name: "Electronics+", Description: "Updated description"}
	rec = doRequest(t, http.MethodPut, "/categories/"+idStr, update, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("update category status = %d, want %d", rec.Code, http.StatusOK)
	}
	var updated models.Category
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.ID != targetID || updated.Name != update.Name {
		t.Fatalf("updated category = %+v, want id=%d name=%q", updated, targetID, update.Name)
	}

	// DELETE
	rec = doRequest(t, http.MethodDelete, "/categories/"+idStr, nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete category status = %d, want %d", rec.Code, http.StatusOK)
	}

	// GET after delete → 404
	rec = doRequest(t, http.MethodGet, "/categories/"+idStr, nil, handler)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted category status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }
