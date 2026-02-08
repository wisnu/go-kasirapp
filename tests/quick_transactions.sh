#!/bin/bash
# Quick script to create a few sample transactions
# Usage: ./tests/quick_transactions.sh [BASE_URL]

BASE="${1:-http://localhost:8080}"

echo "Creating sample transactions..."
echo ""

# Transaction 1: Buy 2 Laptops
echo "1. Buying 2x Laptop..."
curl -s -X POST "$BASE/api/checkout" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":1,"quantity":2}]}' | jq -c '{id, total_amount, items: (.details | length)}'

# Transaction 2: Buy 5 Smartphones  
echo "2. Buying 5x Smartphone..."
curl -s -X POST "$BASE/api/checkout" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":2,"quantity":5}]}' | jq -c '{id, total_amount, items: (.details | length)}'

# Transaction 3: Mixed order
echo "3. Buying 1x Laptop + 3x Tablet + 2x Headphones..."
curl -s -X POST "$BASE/api/checkout" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":1,"quantity":1},{"product_id":3,"quantity":3},{"product_id":4,"quantity":2}]}' | jq -c '{id, total_amount, items: (.details | length)}'

echo ""
echo "Done! Checking today's report..."
echo ""
curl -s "$BASE/api/report/hari-ini" | jq
