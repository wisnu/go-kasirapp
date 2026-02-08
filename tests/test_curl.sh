#!/bin/bash
# Curl-based manual test for kasir-app API
# Jalankan setelah server sudah running: go run kasir-app.go

BASE="${1:-http://localhost:8080}"
PASS=0
FAIL=0
FAILURES=()

# ---------------------------------------------------------------------------
# Helper: print header, run curl, print response
# ---------------------------------------------------------------------------
run() {
    local label="$1"
    shift
    echo ""
    echo "=== $label ==="
    echo "CMD: curl $*"
    # -s  : silent progress
    # -w  : print HTTP status code at the end
    # -o  : output body to stdout via /dev/stdout
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$@")
    BODY=$(echo "$RESPONSE" | sed '$d')
    STATUS=$(echo "$RESPONSE" | tail -1 | sed 's/HTTP_STATUS://')
    echo "STATUS: $STATUS"
    echo "BODY:   $BODY"
}

assert_status() {
    local label="$1"
    local expected="$2"
    if [ "$STATUS" = "$expected" ]; then
        echo "PASS: $label (status $STATUS)"
        PASS=$((PASS + 1))
    else
        echo "FAIL: $label — expected $expected, got $STATUS"
        FAILURES+=("$label — expected status $expected, got $STATUS | body: $BODY")
        FAIL=$((FAIL + 1))
    fi
}

assert_contains() {
    local label="$1"
    local needle="$2"
    if echo "$BODY" | grep -q "$needle"; then
        echo "PASS: $label (contains '$needle')"
        PASS=$((PASS + 1))
    else
        echo "FAIL: $label — body does not contain '$needle'"
        FAILURES+=("$label — expected body to contain '$needle' | body: $BODY")
        FAIL=$((FAIL + 1))
    fi
}

# ===========================================================================
# 1. ROOT & HEALTH
# ===========================================================================
run "GET /" \
    "$BASE/"
assert_status "root returns 200" "200"
assert_contains "root body" "Hello, Ini Backend Program Kasir!"

run "GET /health" \
    "$BASE/health"
assert_status "health returns 200" "200"
assert_contains "health body" '"status":"ok"'

# ===========================================================================
# 2. CATEGORIES — CRUD
# ===========================================================================

# --- GET all categories (seed data: Electronics, Accessories) ---
run "GET /categories" \
    "$BASE/categories"
assert_status "list categories returns 200" "200"
assert_contains "list categories has Electronics" "Electronics"
assert_contains "list categories has Accessories" "Accessories"

# --- POST category baru ---
run "POST /categories" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"name":"Office","description":"Office equipment"}' \
    "$BASE/categories"
assert_status "create category returns 201" "201"
assert_contains "created category name" "Office"

# Ambil ID dari response (cari "id": N)
NEW_CAT_ID=$(echo "$BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
echo "INFO: new category id = $NEW_CAT_ID"

# --- GET category by ID ---
run "GET /categories/$NEW_CAT_ID" \
    "$BASE/categories/$NEW_CAT_ID"
assert_status "get category by id returns 200" "200"
assert_contains "get category body" "Office"

# --- PUT category ---
run "PUT /categories/$NEW_CAT_ID" \
    -X PUT \
    -H "Content-Type: application/json" \
    -d '{"name":"Office Updated","description":"Updated office equipment"}' \
    "$BASE/categories/$NEW_CAT_ID"
assert_status "update category returns 200" "200"
assert_contains "updated category name" "Office Updated"

# --- DELETE category ---
run "DELETE /categories/$NEW_CAT_ID" \
    -X DELETE \
    "$BASE/categories/$NEW_CAT_ID"
assert_status "delete category returns 200" "200"
assert_contains "delete category message" "Category deleted"

# --- GET deleted category → 404 ---
run "GET /categories/$NEW_CAT_ID (deleted)" \
    "$BASE/categories/$NEW_CAT_ID"
assert_status "get deleted category returns 404" "404"

# ===========================================================================
# 3. PRODUCTS — CRUD
# ===========================================================================

# --- GET all products (seed data: 4 produk) ---
run "GET /api/products" \
    "$BASE/api/products"
assert_status "list products returns 200" "200"
assert_contains "list products has Laptop" "Laptop"
assert_contains "list products has category_name" "category_name"

# --- POST product baru (category_id 1 = Electronics dari seed) ---
run "POST /api/products" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"name":"Mouse","price":25.50,"stock":100,"category_id":1}' \
    "$BASE/api/products"
assert_status "create product returns 201" "201"
assert_contains "created product name" "Mouse"
assert_contains "created product category_id" '"category_id":1'

NEW_PROD_ID=$(echo "$BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
echo "INFO: new product id = $NEW_PROD_ID"

# --- GET product by ID ---
run "GET /api/products/$NEW_PROD_ID" \
    "$BASE/api/products/$NEW_PROD_ID"
assert_status "get product by id returns 200" "200"
assert_contains "get product body" "Mouse"
assert_contains "get product has category_name" "category_name"

# --- PUT product (pindah ke category_id 2 = Accessories) ---
run "PUT /api/products/$NEW_PROD_ID" \
    -X PUT \
    -H "Content-Type: application/json" \
    -d '{"name":"Wireless Mouse","price":35.00,"stock":80,"category_id":2}' \
    "$BASE/api/products/$NEW_PROD_ID"
assert_status "update product returns 200" "200"
assert_contains "updated product name" "Wireless Mouse"
assert_contains "updated product category_id" '"category_id":2'

# --- DELETE product ---
run "DELETE /api/products/$NEW_PROD_ID" \
    -X DELETE \
    "$BASE/api/products/$NEW_PROD_ID"
assert_status "delete product returns 200" "200"
assert_contains "delete product message" "Product deleted"

# --- GET deleted product → 404 ---
run "GET /api/products/$NEW_PROD_ID (deleted)" \
    "$BASE/api/products/$NEW_PROD_ID"
assert_status "get deleted product returns 404" "404"

# ===========================================================================
# 4. EDGE CASES
# ===========================================================================

# --- GET product with invalid ID ---
run "GET /api/products/abc" \
    "$BASE/api/products/abc"
assert_status "invalid product id returns 400" "400"

# --- GET category that does not exist ---
run "GET /categories/9999" \
    "$BASE/categories/9999"
assert_status "non-existent category returns 404" "404"

# --- POST with invalid JSON ---
run "POST /api/products (invalid JSON)" \
    -X POST \
    -H "Content-Type: application/json" \
    -d 'not-json' \
    "$BASE/api/products"
assert_status "invalid json returns 400" "400"

# --- POST product with non-existent category_id (FK violation) ---
run "POST /api/products (invalid category_id)" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"name":"Ghost","price":10.00,"stock":1,"category_id":9999}' \
    "$BASE/api/products"
assert_status "invalid category_id returns 500" "500"

# ===========================================================================
# 5. PRODUCT SEARCH
# ===========================================================================

# --- Search products by name (should find "Laptop") ---
run "GET /api/products?name=Lap" \
    "$BASE/api/products?name=Lap"
assert_status "search products returns 200" "200"
assert_contains "search results contain Laptop" "Laptop"

# --- Search products with no results ---
run "GET /api/products?name=NonExistent" \
    "$BASE/api/products?name=NonExistent"
assert_status "search products (no results) returns 200" "200"
# Body should be empty array []
assert_contains "search results empty array" "[]"

# ===========================================================================
# 6. CATEGORY SEARCH
# ===========================================================================

# --- Search categories by name (should find "Electronics") ---
run "GET /categories?name=Elec" \
    "$BASE/categories?name=Elec"
assert_status "search categories returns 200" "200"
assert_contains "search results contain Electronics" "Electronics"

# --- Search categories with no results ---
run "GET /categories?name=NonExistent" \
    "$BASE/categories?name=NonExistent"
assert_status "search categories (no results) returns 200" "200"
assert_contains "search results empty array" "[]"

# ===========================================================================
# 7. TRANSACTIONS / CHECKOUT
# ===========================================================================

# --- POST checkout (create transaction) ---
# Assuming product id 1 (Laptop) and id 2 (Smartphone) exist from seed data
run "POST /api/checkout" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":1,"quantity":1},{"product_id":2,"quantity":2}]}' \
    "$BASE/api/checkout"
assert_status "checkout returns 201" "201"
assert_contains "transaction has total_amount" "total_amount"
assert_contains "transaction has details" "details"
assert_contains "transaction detail has product_name" "product_name"

# Ambil transaction ID dari response
TRANSACTION_ID=$(echo "$BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
echo "INFO: transaction id = $TRANSACTION_ID"

# --- GET all transactions ---
run "GET /api/transactions" \
    "$BASE/api/transactions"
assert_status "list transactions returns 200" "200"
assert_contains "transactions list has created_at" "created_at"

# --- GET transaction by ID ---
run "GET /api/transactions/$TRANSACTION_ID" \
    "$BASE/api/transactions/$TRANSACTION_ID"
assert_status "get transaction by id returns 200" "200"
assert_contains "transaction has details array" "details"
assert_contains "transaction detail has quantity" "quantity"
assert_contains "transaction detail has subtotal" "subtotal"

# --- POST checkout with empty items (should fail) ---
run "POST /api/checkout (empty items)" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"items":[]}' \
    "$BASE/api/checkout"
assert_status "checkout empty items returns 400" "400"
assert_contains "error message empty items" "Items cannot be empty"

# --- POST checkout with invalid product_id (should fail) ---
run "POST /api/checkout (invalid product)" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":9999,"quantity":1}]}' \
    "$BASE/api/checkout"
assert_status "checkout invalid product returns 400" "400"
assert_contains "error message product not found" "not found"

# --- POST checkout with excessive quantity (insufficient stock) ---
run "POST /api/checkout (insufficient stock)" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":1,"quantity":99999}]}' \
    "$BASE/api/checkout"
assert_status "checkout insufficient stock returns 400" "400"
assert_contains "error message insufficient stock" "insufficient stock"

# ===========================================================================
# 8. REPORTS
# ===========================================================================

# --- GET today's report ---
run "GET /api/report/hari-ini" \
    "$BASE/api/report/hari-ini"
assert_status "today report returns 200" "200"
assert_contains "today report has total_revenue" "total_revenue"
assert_contains "today report has total_transaksi" "total_transaksi"
assert_contains "today report has produk_terlaris" "produk_terlaris"
assert_contains "today report has nama" "nama"
assert_contains "today report has qty_terjual" "qty_terjual"

# --- GET report by date range ---
run "GET /api/report?start_date=2026-01-01&end_date=2026-12-31" \
    "$BASE/api/report?start_date=2026-01-01&end_date=2026-12-31"
assert_status "date range report returns 200" "200"
assert_contains "date range report has total_revenue" "total_revenue"
assert_contains "date range report has total_transaksi" "total_transaksi"
assert_contains "date range report has produk_terlaris" "produk_terlaris"

# --- GET report missing start_date ---
run "GET /api/report?end_date=2026-12-31" \
    "$BASE/api/report?end_date=2026-12-31"
assert_status "report missing start_date returns 400" "400"
assert_contains "error message start_date required" "start_date parameter is required"

# --- GET report missing end_date ---
run "GET /api/report?start_date=2026-01-01" \
    "$BASE/api/report?start_date=2026-01-01"
assert_status "report missing end_date returns 400" "400"
assert_contains "error message end_date required" "end_date parameter is required"

# --- GET report with invalid date format ---
run "GET /api/report?start_date=01-01-2026&end_date=2026-12-31" \
    "$BASE/api/report?start_date=01-01-2026&end_date=2026-12-31"
assert_status "report invalid date format returns 400" "400"
assert_contains "error message invalid format" "Invalid date format"

# ===========================================================================
# SUMMARY
# ===========================================================================
echo ""
echo "================================================"
echo "  TOTAL: $((PASS + FAIL))  |  PASS: $PASS  |  FAIL: $FAIL"
echo "================================================"

if [ "$FAIL" -gt 0 ]; then
    echo ""
    echo "--- FAILED TESTS ---"
    for i in "${!FAILURES[@]}"; do
        echo "  $((i + 1)). ${FAILURES[$i]}"
    done
    echo "--------------------"
    exit 1
fi
