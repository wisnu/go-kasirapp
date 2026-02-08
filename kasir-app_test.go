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

func TestProductsSearch(t *testing.T) {
	var handler http.HandlerFunc
	
	if isIntegration() {
		// In integration mode, we'll search against real data
		// No setup needed, just use nil handler
		handler = nil
	} else {
		// Unit mode: setup mock
		h, mock := setupProductHandler(t)
		handler = h.Handle

		// Mock search results for "Lap" (should match "Laptop")
		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "category_name"}).
			AddRow(1, "Laptop", 999.99, 10, "Electronics").
			AddRow(5, "Laptop Pro", 1299.99, 5, "Electronics")
		mock.ExpectQuery("SELECT p.id, p.name, p.price, p.stock, c.name").
			WithArgs("%Lap%").
			WillReturnRows(rows)

		// Mock search with no results
		mock.ExpectQuery("SELECT p.id, p.name, p.price, p.stock, c.name").
			WithArgs("%NonExistent%").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "stock", "category_name"}))

		defer func() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		}()
	}

	// Test 1: Search with results
	rec := doRequest(t, http.MethodGet, "/api/products?name=Lap", nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("search products status = %d, want %d", rec.Code, http.StatusOK)
	}
	var results []models.Product
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode search results: %v", err)
	}
	
	// In unit mode, we expect exactly 2 results
	if !isIntegration() && len(results) != 2 {
		t.Fatalf("search results len = %d, want 2", len(results))
	}
	
	// In integration mode, we just verify we got a valid response
	if isIntegration() {
		t.Logf("integration mode: found %d products matching 'Lap'", len(results))
	} else {
		// Verify the results contain "Laptop" in the name
		for _, p := range results {
			if p.Name != "Laptop" && p.Name != "Laptop Pro" {
				t.Fatalf("unexpected product in search results: %s", p.Name)
			}
		}
	}

	// Test 2: Search with no results
	rec = doRequest(t, http.MethodGet, "/api/products?name=NonExistent", nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("search products (no results) status = %d, want %d", rec.Code, http.StatusOK)
	}
	var emptyResults []models.Product
	if err := json.NewDecoder(rec.Body).Decode(&emptyResults); err != nil {
		t.Fatalf("decode empty search results: %v", err)
	}
	
	if !isIntegration() && len(emptyResults) != 0 {
		t.Fatalf("search (no results) len = %d, want 0", len(emptyResults))
	}
}

func TestCategoriesSearch(t *testing.T) {
	var handler http.HandlerFunc
	
	if isIntegration() {
		// In integration mode, we'll search against real data
		handler = nil
	} else {
		// Unit mode: setup mock
		h, mock := setupCategoryHandler(t)
		handler = h.Handle

		// Mock search results for "Elec" (should match "Electronics")
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).
			AddRow(1, "Electronics", "Electronic devices and gadgets").
			AddRow(3, "Electronic Accessories", "Accessories for electronic devices")
		mock.ExpectQuery("SELECT id, name, description FROM categories WHERE name ILIKE").
			WithArgs("%Elec%").
			WillReturnRows(rows)

		// Mock search with no results
		mock.ExpectQuery("SELECT id, name, description FROM categories WHERE name ILIKE").
			WithArgs("%NonExistent%").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}))

		defer func() {
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		}()
	}

	// Test 1: Search with results
	rec := doRequest(t, http.MethodGet, "/categories?name=Elec", nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("search categories status = %d, want %d", rec.Code, http.StatusOK)
	}
	var results []models.Category
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode search results: %v", err)
	}
	
	// In unit mode, we expect exactly 2 results
	if !isIntegration() && len(results) != 2 {
		t.Fatalf("search results len = %d, want 2", len(results))
	}
	
	// In integration mode, we just verify we got a valid response
	if isIntegration() {
		t.Logf("integration mode: found %d categories matching 'Elec'", len(results))
	} else {
		// Verify the results contain "Elec" in the name
		for _, c := range results {
			if c.Name != "Electronics" && c.Name != "Electronic Accessories" {
				t.Fatalf("unexpected category in search results: %s", c.Name)
			}
		}
	}

	// Test 2: Search with no results
	rec = doRequest(t, http.MethodGet, "/categories?name=NonExistent", nil, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("search categories (no results) status = %d, want %d", rec.Code, http.StatusOK)
	}
	var emptyResults []models.Category
	if err := json.NewDecoder(rec.Body).Decode(&emptyResults); err != nil {
		t.Fatalf("decode empty search results: %v", err)
	}
	
	if !isIntegration() && len(emptyResults) != 0 {
		t.Fatalf("search (no results) len = %d, want 0", len(emptyResults))
	}
}

func TestTodayReport(t *testing.T) {
	// This endpoint doesn't require mocking complex queries in unit mode,
	// so we'll primarily test in integration mode
	// In unit mode, we'd need to mock transaction queries which is complex
	
	if !isIntegration() {
		t.Skip("Skipping today report test in unit mode (requires complex transaction mocking)")
	}
	
	// GET today's report
	rec := doRequest(t, http.MethodGet, "/api/report/hari-ini", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("today report status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var report models.DailyReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode today report: %v", err)
	}
	
	// Verify structure (values may be 0 if no transactions today)
	t.Logf("Today report: revenue=%d, transactions=%d, best_product=%s (qty=%d)",
		report.TotalRevenue, report.TotalTransaksi, report.ProdukTerlaris.Nama, report.ProdukTerlaris.QtyTerjual)
}

func TestReportByDateRange(t *testing.T) {
	if !isIntegration() {
		t.Skip("Skipping date range report test in unit mode (requires complex transaction mocking)")
	}
	
	// Test 1: Valid date range
	rec := doRequest(t, http.MethodGet, "/api/report?start_date=2026-01-01&end_date=2026-12-31", nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("date range report status = %d, want %d", rec.Code, http.StatusOK)
	}
	
	var report models.DailyReport
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode date range report: %v", err)
	}
	
	t.Logf("Date range report (2026): revenue=%d, transactions=%d, best_product=%s (qty=%d)",
		report.TotalRevenue, report.TotalTransaksi, report.ProdukTerlaris.Nama, report.ProdukTerlaris.QtyTerjual)
	
	// Test 2: Missing start_date
	rec = doRequest(t, http.MethodGet, "/api/report?end_date=2026-12-31", nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("report missing start_date status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	
	// Test 3: Missing end_date
	rec = doRequest(t, http.MethodGet, "/api/report?start_date=2026-01-01", nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("report missing end_date status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	
	// Test 4: Invalid date format
	rec = doRequest(t, http.MethodGet, "/api/report?start_date=01-01-26&end_date=2026-12-31", nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("report invalid date format status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }
