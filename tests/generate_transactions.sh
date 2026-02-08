#!/bin/bash
# Script to generate sample transactions with varied products
# Usage: ./tests/generate_transactions.sh [BASE_URL]

BASE="${1:-http://localhost:8080}"
TOTAL_TRANSACTIONS=10
SUCCESS_COUNT=0
FAIL_COUNT=0

echo "=========================================="
echo "  Generating Sample Transactions"
echo "=========================================="
echo "Base URL: $BASE"
echo "Target: $TOTAL_TRANSACTIONS transactions"
echo ""

# Helper function to create transaction
create_transaction() {
    local num=$1
    local items=$2
    local description=$3
    
    echo "[$num/$TOTAL_TRANSACTIONS] Creating transaction: $description"
    
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$items" \
        "$BASE/api/checkout")
    
    BODY=$(echo "$RESPONSE" | sed '$d')
    STATUS=$(echo "$RESPONSE" | tail -1 | sed 's/HTTP_STATUS://')
    
    if [ "$STATUS" = "201" ]; then
        TRANSACTION_ID=$(echo "$BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
        TOTAL_AMOUNT=$(echo "$BODY" | grep -o '"total_amount":[0-9]*' | head -1 | cut -d: -f2)
        echo "  ✓ SUCCESS - Transaction ID: $TRANSACTION_ID, Total: Rp $TOTAL_AMOUNT"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        ERROR=$(echo "$BODY" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
        echo "  ✗ FAILED - Status: $STATUS, Error: $ERROR"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    echo ""
    sleep 0.5
}

# Transaction 1: Single Laptop
create_transaction 1 \
    '{"items":[{"product_id":1,"quantity":1}]}' \
    "1x Laptop"

# Transaction 2: Multiple Smartphones
create_transaction 2 \
    '{"items":[{"product_id":2,"quantity":3}]}' \
    "3x Smartphone"

# Transaction 3: Tablet only
create_transaction 3 \
    '{"items":[{"product_id":3,"quantity":2}]}' \
    "2x Tablet"

# Transaction 4: Headphones bundle
create_transaction 4 \
    '{"items":[{"product_id":4,"quantity":5}]}' \
    "5x Headphones"

# Transaction 5: Mix - Laptop + Smartphone
create_transaction 5 \
    '{"items":[{"product_id":1,"quantity":1},{"product_id":2,"quantity":2}]}' \
    "1x Laptop + 2x Smartphone"

# Transaction 6: Mix - All electronics
create_transaction 6 \
    '{"items":[{"product_id":1,"quantity":1},{"product_id":2,"quantity":1},{"product_id":3,"quantity":1}]}' \
    "1x Laptop + 1x Smartphone + 1x Tablet"

# Transaction 7: Smartphone + Headphones combo
create_transaction 7 \
    '{"items":[{"product_id":2,"quantity":2},{"product_id":4,"quantity":2}]}' \
    "2x Smartphone + 2x Headphones"

# Transaction 8: Tablet + Headphones
create_transaction 8 \
    '{"items":[{"product_id":3,"quantity":1},{"product_id":4,"quantity":3}]}' \
    "1x Tablet + 3x Headphones"

# Transaction 9: Large order - Multiple items
create_transaction 9 \
    '{"items":[{"product_id":2,"quantity":4},{"product_id":3,"quantity":2},{"product_id":4,"quantity":6}]}' \
    "4x Smartphone + 2x Tablet + 6x Headphones"

# Transaction 10: Premium combo - Laptop + accessories
create_transaction 10 \
    '{"items":[{"product_id":1,"quantity":2},{"product_id":4,"quantity":4}]}' \
    "2x Laptop + 4x Headphones"

echo "=========================================="
echo "  Transaction Generation Summary"
echo "=========================================="
echo "Total Attempted: $TOTAL_TRANSACTIONS"
echo "Successful:      $SUCCESS_COUNT"
echo "Failed:          $FAIL_COUNT"
echo ""

# Get today's report
echo "=========================================="
echo "  Today's Report"
echo "=========================================="
curl -s "$BASE/api/report/hari-ini" | python3 -m json.tool 2>/dev/null || curl -s "$BASE/api/report/hari-ini" | jq 2>/dev/null || curl -s "$BASE/api/report/hari-ini"
echo ""

# Get recent transactions
echo "=========================================="
echo "  Recent Transactions (Last 5)"
echo "=========================================="
TRANSACTIONS=$(curl -s "$BASE/api/transactions")
echo "$TRANSACTIONS" | python3 -m json.tool 2>/dev/null | head -30 || echo "$TRANSACTIONS" | jq '.[0:5]' 2>/dev/null || echo "$TRANSACTIONS"
echo ""

echo "=========================================="
echo "  Done!"
echo "=========================================="
