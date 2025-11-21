package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// Test configuration
var (
	userServiceURL     = getEnv("USER_SERVICE_URL", "http://localhost:8181")
	expensesServiceURL = getEnv("EXPENSES_SERVICE_URL", "http://localhost:8282")
	paymentServiceURL  = getEnv("PAYMENT_SERVICE_URL", "http://localhost:8989")
	bankServiceURL     = getEnv("BANK_SERVICE_URL", "http://localhost:8585")
)

type User struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserId string `json:"userId"`
}

type Expense struct {
	ExpenseId string  `json:"expenseId"`
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

type Transaction struct {
	TransactionId string  `json:"transactionId"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestFullUserWorkflow tests the complete user journey:
// 1. Register user
// 2. Login
// 3. Create expense
// 4. Process payment
// 5. Verify transaction
func TestFullUserWorkflow(t *testing.T) {
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("e2e_user_%d", timestamp)
	email := fmt.Sprintf("e2e_%d@example.com", timestamp)
	password := "testpass123"

	// Step 1: Register user
	t.Log("Step 1: Registering user...")
	userData := map[string]interface{}{
		"firstName": "E2E",
		"lastName":  "Test",
		"username":  username,
		"email":     email,
		"password":  password,
	}

	jsonData, _ := json.Marshal(userData)
	resp, err := http.Post(userServiceURL+"/user", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Registration failed. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode user: %v", err)
	}

	if user.UserId == "" {
		t.Fatal("User ID is empty")
	}
	t.Logf("✓ User registered: %s", user.UserId)

	// Step 2: Login
	t.Log("Step 2: Logging in...")
	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, _ = json.Marshal(loginData)
	resp, err = http.Post(userServiceURL+"/user/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Login failed. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if loginResp.Token == "" {
		t.Fatal("Token is empty")
	}
	t.Logf("✓ Login successful. Token received.")

	// Step 3: Create expense
	t.Log("Step 3: Creating expense...")
	expenseData := map[string]interface{}{
		"name":        "E2E Test Expense",
		"description": "End-to-end test expense",
		"amount":      75.50,
		"category":    "testing",
		"tags":        []string{"e2e", "test"},
	}

	jsonData, _ = json.Marshal(expenseData)
	req, _ := http.NewRequest("POST", expensesServiceURL+"/expenses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create expense: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Failed to create expense. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

	var expense Expense
	if err := json.NewDecoder(resp.Body).Decode(&expense); err != nil {
		t.Fatalf("Failed to decode expense: %v", err)
	}

	if expense.ExpenseId == "" {
		t.Fatal("Expense ID is empty")
	}
	t.Logf("✓ Expense created: %s (Amount: %.2f)", expense.ExpenseId, expense.Amount)

	// Step 4: Process payment
	t.Log("Step 4: Processing payment...")
	paymentData := map[string]string{
		"expenseId": expense.ExpenseId,
	}

	jsonData, _ = json.Marshal(paymentData)
	req, _ = http.NewRequest("POST", paymentServiceURL+"/payment", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)

	client = &http.Client{Timeout: 15 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to process payment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Payment processing failed. Status: %d, Body: %s", resp.StatusCode, string(body))
	}

	var transaction Transaction
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		t.Fatalf("Failed to decode transaction: %v", err)
	}

	if transaction.TransactionId == "" {
		t.Fatal("Transaction ID is empty")
	}
	t.Logf("✓ Payment processed: %s (Status: %s)", transaction.TransactionId, transaction.Status)

	// Step 5: Verify transaction
	t.Log("Step 5: Verifying transaction...")
	req, _ = http.NewRequest("GET", paymentServiceURL+"/payment/"+transaction.TransactionId, nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)

	client = &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get transaction. Status: %d", resp.StatusCode)
	}

	var verifyTransaction Transaction
	if err := json.NewDecoder(resp.Body).Decode(&verifyTransaction); err != nil {
		t.Fatalf("Failed to decode transaction: %v", err)
	}

	if verifyTransaction.TransactionId != transaction.TransactionId {
		t.Fatalf("Transaction ID mismatch. Expected: %s, Got: %s", transaction.TransactionId, verifyTransaction.TransactionId)
	}
	t.Logf("✓ Transaction verified: %s", verifyTransaction.TransactionId)

	t.Log("✓ Full workflow completed successfully!")
}
