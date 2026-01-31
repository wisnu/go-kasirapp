#!/bin/bash
# Curl-based manual test for kasir-app API
# Jalankan setelah server sudah running: go run kasir-app.go

set -e

BASE="http://localhost:8080"
PASS=0
FAIL=0

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

# ===========================================================================
# SUMMARY
# ===========================================================================
echo ""
echo "================================================"
echo "  TOTAL: $((PASS + FAIL))  |  PASS: $PASS  |  FAIL: $FAIL"
echo "================================================"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
