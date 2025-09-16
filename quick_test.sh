#!/bin/bash

# Quick Manual Testing Script for TrustlessWork Indexer
set -e

echo "🧪 Quick Manual Testing Script"
echo "============================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m' 
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to run test and show result
run_test() {
    local test_name="$1"
    local curl_command="$2"
    
    echo -e "\n${BLUE}🔍 Test: $test_name${NC}"
    echo "Command: $curl_command"
    echo -e "${YELLOW}Response:${NC}"
    
    if eval "$curl_command"; then
        echo -e "\n${GREEN}✅ Test passed${NC}"
    else
        echo -e "\n${RED}❌ Test failed${NC}"
    fi
    
    echo "----------------------------------------"
}

# Check if server is running
if ! curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${RED}❌ Server is not running on port 8080${NC}"
    echo "Please start the server first:"
    echo "  go build -o indexer cmd/indexer/main.go"
    echo "  ./indexer"
    exit 1
fi

echo -e "${GREEN}✅ Server is running${NC}"

# Test 1: Create Single Release Escrow
run_test "Create Single Release Escrow (Basic)" \
    "curl -s -X POST http://localhost:8080/escrows/single -H 'Content-Type: application/json' -d @tests/single_basic.json"

# Test 2: Get the created escrow
run_test "Get Single Release Escrow" \
    "curl -s -X GET http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF | jq '.contractId, .amount.raw, .milestones[0].description'"

# Test 3: Create Multi Release Escrow
run_test "Create Multi Release Escrow (Basic)" \
    "curl -s -X POST http://localhost:8080/escrows/multi -H 'Content-Type: application/json' -d @tests/multi_basic.json"

# Test 4: Get multi release escrow
run_test "Get Multi Release Escrow with Milestones" \
    "curl -s -X GET http://localhost:8080/escrows/MULTI001BASICTEST123456789ABCDEF | jq '.contractId, .milestones | length, .totalAmount.raw'"

# Test 5: Index Deposits
run_test "Index Funder Deposits (Mock Data)" \
    "curl -s -X POST http://localhost:8080/index/funder-deposits/MULTI001BASICTEST123456789ABCDEF | jq '.ok, .deposits | length'"

# Test 6: Test High Amount Single Release
run_test "Create High Amount Single Release" \
    "curl -s -X POST http://localhost:8080/escrows/single -H 'Content-Type: application/json' -d @tests/single_high_amount.json"

# Test 7: Verify high amount handling
run_test "Verify High Amount Handling" \
    "curl -s -X GET http://localhost:8080/escrows/SINGLE002HIGHAMOUNT123456789ABCDEF | jq '.amount.raw'"

# Test 8: Complex Multi Release
run_test "Create Complex Multi Release (7 milestones)" \
    "curl -s -X POST http://localhost:8080/escrows/multi -H 'Content-Type: application/json' -d @tests/multi_complex.json"

# Test 9: Verify complex multi release
run_test "Verify Complex Multi Release Structure" \
    "curl -s -X GET http://localhost:8080/escrows/MULTI002COMPLEX123456789ABCDEF | jq '.milestones | length, .totalAmount.raw'"

# Test 10: Delete Operations
run_test "Delete Single Release Escrow" \
    "curl -s -X DELETE http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF"

run_test "Delete Multi Release Escrow" \
    "curl -s -X DELETE http://localhost:8080/escrows/MULTI001BASICTEST123456789ABCDEF"

# Test 11: Verify deletion
echo -e "\n${BLUE}🔍 Test: Verify Deletion (should return 404 or error)${NC}"
echo "Command: curl -s -X GET http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF"
echo -e "${YELLOW}Response:${NC}"
if curl -s -X GET http://localhost:8080/escrows/SINGLE001BASICTEST123456789ABCDEF | grep -q "not found\|error"; then
    echo -e "\n${GREEN}✅ Deletion verified - escrow not found (expected)${NC}"
else
    echo -e "\n${YELLOW}⚠️  Escrow might still exist or different error occurred${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Quick testing completed!${NC}"
echo ""
echo "For more detailed testing, see:"
echo "  - manual_test_guide.md"
echo "  - Individual JSON files in tests/ directory"
echo "  - Database verification commands in the guide"