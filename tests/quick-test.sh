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
    USER_PORT=""
    EXPENSE_PORT=""
    PAYMENT_PORT=""
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
USERNAME="testuser${TIMESTAMP}"  # Alphanumeric only (no underscores)
EMAIL="test${TIMESTAMP}@example.com"
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

    echo -e "${BLUE}📝 $step_name${NC}" >&2
    echo -e "   Request: ${method} ${endpoint}" >&2

    local start_time=$(date +%s%N)
    local curl_args=(-k -s -w "\nHTTP_STATUS:%{http_code}" --max-time 30 -X "$method" "${BASE_URL}${endpoint}" -H "Content-Type: application/json")

    # Add auth header if provided
    if [[ -n "$auth_header" ]]; then
        curl_args+=(-H "Authorization: Bearer ${auth_header}")
    fi

    # Add data only if provided (for POST/PUT requests)
    if [[ -n "$data" ]]; then
        curl_args+=(-d "$data")
    fi

    response=$(curl "${curl_args[@]}" 2>&1)
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))

    # Extract status code
    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2 || echo "000")
    body=$(echo "$response" | sed '/HTTP_STATUS:/d')
    
    # Handle curl errors (timeout, connection refused, etc.)
    if [[ -z "$http_status" || "$http_status" == "000" ]]; then
        echo -e "   ${RED}❌ Connection Error${NC}" >&2
        echo -e "   Response: $response" >&2
        echo "" >&2
        return 1
    fi

    # Check if successful
    if [[ "$http_status" =~ ^2[0-9][0-9]$ ]]; then
        echo -e "   ${GREEN}✅ Status: $http_status${NC}" >&2
        echo -e "   ⏱️  Response Time: ${duration}ms" >&2
        echo "$body" | jq '.' 2>/dev/null || echo "$body" >&2
    else
        echo -e "   ${RED}❌ Status: $http_status${NC}" >&2
        echo -e "   ⏱️  Response Time: ${duration}ms" >&2
        echo "$body" | jq '.' 2>/dev/null || echo "$body" >&2
        echo "" >&2
        return 1
    fi

    echo "" >&2
    # Return the body so it can be captured (to stdout)
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
    USER_ID=$(echo "$result" | jq -r '.userId // empty')
    if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
        echo -e "   ${GREEN}User ID: $USER_ID${NC}"
    fi
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
    TOKEN=$(echo "$result" | jq -r '.token // empty')
    if [[ -n "$TOKEN" && "$TOKEN" != "null" ]]; then
        echo -e "   ${GREEN}Token received${NC}"
    fi
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
    EXPENSE_ID=$(echo "$result" | jq -r '.expenseId // empty')
    if [[ -n "$EXPENSE_ID" && "$EXPENSE_ID" != "null" ]]; then
        echo -e "   ${GREEN}Expense ID: $EXPENSE_ID${NC}"
    fi
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
    TRANSACTION_ID=$(echo "$result" | jq -r '.transactionId // empty')
    if [[ -n "$TRANSACTION_ID" && "$TRANSACTION_ID" != "null" ]]; then
        echo -e "   ${GREEN}Transaction ID: $TRANSACTION_ID${NC}"
    fi
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
