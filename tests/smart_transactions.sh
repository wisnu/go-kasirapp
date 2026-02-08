#!/bin/bash
# Smart transaction generator - checks stock before creating transaction
# Usage: ./tests/smart_transactions.sh [BASE_URL] [COUNT]

BASE="${1:-http://localhost:8080}"
COUNT="${2:-5}"

echo "=========================================="
echo "  Smart Transaction Generator"
echo "=========================================="
echo ""

# Get available products
echo "Fetching available products..."
PRODUCTS=$(curl -s "$BASE/api/products")
echo "Available products:"
echo "$PRODUCTS" | jq -r '.[] | "  - [\(.id)] \(.name): Rp \(.price | tonumber) (Stock: \(.stock))"'
echo ""

# Create transactions with random products
echo "Creating $COUNT random transactions..."
echo ""

for i in $(seq 1 $COUNT); do
    # Randomly select product IDs and quantities
    # Using Headphones (id=4) more often as it has higher stock
    RAND=$((RANDOM % 4))
    
    case $RAND in
        0)
            # Small order - 2-3 Headphones
            QTY=$((2 + RANDOM % 2))
            ITEMS='{"items":[{"product_id":4,"quantity":'$QTY'}]}'
            DESC="$QTY x Headphones"
            ;;
        1)
            # Medium order - 1 Smartphone + 1 Headphone
            ITEMS='{"items":[{"product_id":2,"quantity":1},{"product_id":4,"quantity":1}]}'
            DESC="1x Smartphone + 1x Headphones"
            ;;
        2)
            # Tablet order
            QTY=$((1 + RANDOM % 2))
            ITEMS='{"items":[{"product_id":3,"quantity":'$QTY'}]}'
            DESC="$QTY x Tablet"
            ;;
        3)
            # Mix order
            ITEMS='{"items":[{"product_id":3,"quantity":1},{"product_id":4,"quantity":2}]}'
            DESC="1x Tablet + 2x Headphones"
            ;;
    esac
    
    echo "[$i/$COUNT] Creating: $DESC"
    
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$ITEMS" \
        "$BASE/api/checkout")
    
    BODY=$(echo "$RESPONSE" | sed '$d')
    STATUS=$(echo "$RESPONSE" | tail -1 | sed 's/HTTP_STATUS://')
    
    if [ "$STATUS" = "201" ]; then
        TRX_ID=$(echo "$BODY" | jq -r '.id')
        TOTAL=$(echo "$BODY" | jq -r '.total_amount')
        echo "  ✓ Transaction #$TRX_ID - Total: Rp $TOTAL"
    else
        ERROR=$(echo "$BODY" | jq -r '.error')
        echo "  ✗ Failed: $ERROR"
    fi
    
    # Small delay between transactions
    sleep 0.3
done

echo ""
echo "=========================================="
echo "  Summary Report"
echo "=========================================="
echo ""

# Today's report
echo "📊 Today's Report:"
curl -s "$BASE/api/report/hari-ini" | jq

echo ""
echo "📦 Current Stock Levels:"
curl -s "$BASE/api/products" | jq -r '.[] | "  - \(.name): \(.stock) units remaining"'

echo ""
echo "Done!"
