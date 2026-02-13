#!/bin/bash
set -euo pipefail

# Smart-Cash API Quick Test Script
# Usage: ./quick-test.sh [environment] [scenario]
# Examples:
#   ./quick-test.sh dev full       # Full workflow on dev
#   ./quick-test.sh dev register   # Just register user on dev
#   ./quick-test.sh local login    # Login on local

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT="${1:-dev}"
SCENARIO="${2:-full}"

if [[ "$ENVIRONMENT" == "dev" ]]; then
    BASE_URL="https://api.develop.smartcash.danielrive.site"
elif [[ "$ENVIRONMENT" == "local" ]]; then
    BASE_URL="http://localhost"
    USER_PORT=":8181"
    EXPENSE_PORT=":8282"
    PAYMENT_PORT=":8989"
else
    echo "Unknown environment: $ENVIRONMENT (use 'dev' or 'local')"
    exit 1
fi

# Generate unique test data
TIMESTAMP=$(date +%s)
USERNAME="test_user_${TIMESTAMP}"
EMAIL="test_${TIMESTAMP}@example.com"
PASSWORD="testpass123"

# Store test data
USER_ID=""
TOKEN=""
EXPENSE_ID=""
TRANSACTION_ID=""

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 Smart-Cash API Test${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "Environment: ${YELLOW}$ENVIRONMENT${NC}"
echo -e "Scenario: ${YELLOW}$SCENARIO${NC}"
echo -e "Base URL: ${YELLOW}$BASE_URL${NC}"
echo ""

# Function to make requests with timing
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local auth_header=$4
    local step_name=$5

    echo -e "${BLUE}📝 $step_name${NC}"
    echo -e "   Request: ${method} ${endpoint}"

    local start_time=$(date +%s%N)

    if [[ "$auth_header" != "" ]]; then
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" \
            "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${auth_header}" \
            -d "$data")
    else
        response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X "$method" \
            "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))

    # Extract status code
    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_STATUS:/d')

    # Check if successful
    if [[ "$http_status" =~ ^2[0-9][0-9]$ ]]; then
        echo -e "   ${GREEN}✅ Status: $http_status${NC}"
        echo -e "   ⏱️  Response Time: ${duration}ms"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        echo -e "   ${RED}❌ Status: $http_status${NC}"
        echo -e "   ⏱️  Response Time: ${duration}ms"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi

    echo ""
    echo "$body"
}

# Test scenarios
register_user() {
    local payload=$(cat <<EOF
{
  "firstName": "Test",
  "lastName": "User",
  "username": "$USERNAME",
  "email": "$EMAIL",
  "password": "$PASSWORD"
}
EOF
)

    result=$(make_request "POST" "${USER_PORT}/user" "$payload" "" "Step 1: Register User")
    USER_ID=$(echo "$result" | jq -r '.userId')
    echo -e "   ${GREEN}User ID: $USER_ID${NC}"
}

login_user() {
    local payload=$(cat <<EOF
{
  "username": "$USERNAME",
  "password": "$PASSWORD"
}
EOF
)

    result=$(make_request "POST" "${USER_PORT}/user/login" "$payload" "" "Step 2: Login User")
    TOKEN=$(echo "$result" | jq -r '.token')
    echo -e "   ${GREEN}Token received${NC}"
}

create_expense() {
    if [[ -z "$TOKEN" ]]; then
        echo -e "${RED}❌ No token available. Run login first.${NC}"
        return 1
    fi

    local payload=$(cat <<EOF
{
  "name": "Test Expense",
  "description": "Automated test expense",
  "amount": 75.50,
  "category": "testing",
  "tags": ["test", "automated"]
}
EOF
)

    result=$(make_request "POST" "${EXPENSE_PORT}/expenses" "$payload" "$TOKEN" "Step 3: Create Expense")
    EXPENSE_ID=$(echo "$result" | jq -r '.expenseId')
    echo -e "   ${GREEN}Expense ID: $EXPENSE_ID${NC}"
}

process_payment() {
    if [[ -z "$TOKEN" ]]; then
        echo -e "${RED}❌ No token available. Run login first.${NC}"
        return 1
    fi

    if [[ -z "$EXPENSE_ID" ]]; then
        echo -e "${RED}❌ No expense ID available. Create expense first.${NC}"
        return 1
    fi

    local payload=$(cat <<EOF
{
  "expenseId": "$EXPENSE_ID"
}
EOF
)

    result=$(make_request "POST" "${PAYMENT_PORT}/payment" "$payload" "$TOKEN" "Step 4: Process Payment")
    TRANSACTION_ID=$(echo "$result" | jq -r '.transactionId')
    echo -e "   ${GREEN}Transaction ID: $TRANSACTION_ID${NC}"
}

verify_transaction() {
    if [[ -z "$TOKEN" ]]; then
        echo -e "${RED}❌ No token available. Run login first.${NC}"
        return 1
    fi

    if [[ -z "$TRANSACTION_ID" ]]; then
        echo -e "${RED}❌ No transaction ID available. Process payment first.${NC}"
        return 1
    fi

    result=$(make_request "GET" "${PAYMENT_PORT}/payment/$TRANSACTION_ID" "" "$TOKEN" "Step 5: Verify Transaction")
}

# Run scenario
case "$SCENARIO" in
    "register")
        register_user
        ;;
    "login")
        register_user
        login_user
        ;;
    "expense")
        register_user
        login_user
        create_expense
        ;;
    "payment")
        register_user
        login_user
        create_expense
        process_payment
        ;;
    "full")
        register_user
        login_user
        create_expense
        process_payment
        verify_transaction
        ;;
    *)
        echo "Unknown scenario: $SCENARIO"
        echo "Available: register, login, expense, payment, full"
        exit 1
        ;;
esac

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Test completed!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
