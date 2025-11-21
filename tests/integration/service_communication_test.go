package integration

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Test configuration
var (
	userServiceURL     = getEnv("USER_SERVICE_URL", "https://api.develop.smartcash.danielrive.site")
	expensesServiceURL = getEnv("EXPENSES_SERVICE_URL", "https://api.develop.smartcash.danielrive.site")
	paymentServiceURL  = getEnv("PAYMENT_SERVICE_URL", "https://api.develop.smartcash.danielrive.site")
	bankServiceURL     = getEnv("BANK_SERVICE_URL", "https://api.develop.smartcash.danielrive.site")
	testUsername       = getEnv("TEST_USERNAME", "testuser")
	testPassword       = getEnv("TEST_PASSWORD", "testpass123")
	testEmail          = getEnv("TEST_EMAIL", "test@example.com")
)

type User struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
}

type CreateUserResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserId string `json:"userId"`
}

type Expense struct {
	ExpenseId   string  `json:"expenseId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
}

type Transaction struct {
	TransactionId string  `json:"transactionId"`
	ExpenseId     string  `json:"expenseId"`
	UserId        string  `json:"userId"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createHTTPClient creates an HTTP client that skips SSL verification
func createHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
		// Handle redirects - preserve POST method for 307/308, but 301/302 will convert to GET
		// This is standard HTTP behavior, so we need to use the correct URL
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// For 307/308, preserve method. For 301/302, method changes to GET (HTTP standard)
			// Best solution: use correct URL that doesn't redirect
			return nil // Allow redirect, but be aware method may change
		},
	}
}

func TestServiceHealth(t *testing.T) {
	services := map[string]string{
		"user":     userServiceURL + "/user/health",
		"expenses": expensesServiceURL + "/expenses/health",
		"payment":  paymentServiceURL + "/payment/health",
		"bank":     bankServiceURL + "/bank/health",
	}

	client := createHTTPClient()

	for name, url := range services {
		t.Run(name, func(t *testing.T) {
			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("Failed to connect to %s service: %v", name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestUserRegistrationAndLogin(t *testing.T) {
	// Register user
	timestamp := time.Now().Unix()
	userData := map[string]interface{}{
		"firstName": "Test",
		"lastName":  "User",
		"username":  "testuser", // fmt.Sprintf("%d", timestamp),
		"email":     fmt.Sprintf("test_%d@example.com", timestamp),
		"password":  testPassword,
	}
	client := createHTTPClient()

	jsonData, _ := json.Marshal(userData)
	req, err := http.NewRequest("POST", userServiceURL+"/user", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body[:n]))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var createResp CreateUserResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		t.Fatalf("Failed to decode response: %v. Body: %s", err, string(body))
	}

	user := createResp.User
	if user.UserId == "" {
		t.Fatalf("User ID is empty. Response: %s", string(body))
	}

	// Login
	loginData := map[string]string{
		"username": userData["username"].(string),
		"password": testPassword,
	}

	jsonData, _ = json.Marshal(loginData)
	req, err = http.NewRequest("POST", userServiceURL+"/user/login", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if loginResp.Token == "" {
		t.Fatal("Token is empty")
	}

	t.Logf("Successfully registered and logged in user: %s", user.UserId)
}

func TestExpenseServiceCallsUserService(t *testing.T) {
	// This test validates that expenses service can call user service
	// We'll create an expense which should trigger a call to user service

	// First, login to get token
	token := getAuthToken(t)

	// Create expense (this should trigger user validation)
	expenseData := map[string]interface{}{
		"name":        "Test Expense",
		"description": "Integration test expense",
		"amount":      50.0,
		"category":    "test",
		"tags":        []string{"integration-test"},
	}

	jsonData, _ := json.Marshal(expenseData)
	req, _ := http.NewRequest("POST", expensesServiceURL+"/expenses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create expense: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var expense Expense
	if err := json.NewDecoder(resp.Body).Decode(&expense); err != nil {
		t.Fatalf("Failed to decode expense: %v", err)
	}

	if expense.ExpenseId == "" {
		t.Fatal("Expense ID is empty")
	}

	t.Logf("Successfully created expense: %s", expense.ExpenseId)
}

func TestPaymentServiceCallsExpensesService(t *testing.T) {
	// This test validates that payment service can call expenses service

	// Login and create expense first
	token := getAuthToken(t)
	expenseId := createExpense(t, token)

	// Process payment (this should trigger calls to expenses and bank services)
	paymentData := map[string]string{
		"expenseId": expenseId,
	}

	jsonData, _ := json.Marshal(paymentData)
	req, _ := http.NewRequest("POST", paymentServiceURL+"/payment", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := createHTTPClient()
	client.Timeout = 15 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to process payment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		resp.Body.Read(body)
		t.Fatalf("Expected status 201 or 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var transaction Transaction
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		t.Fatalf("Failed to decode transaction: %v", err)
	}

	if transaction.TransactionId == "" {
		t.Fatal("Transaction ID is empty")
	}

	t.Logf("Successfully processed payment: %s", transaction.TransactionId)
}

// Helper functions

func getAuthToken(t *testing.T) string {
	loginData := map[string]string{
		"username": testUsername,
		"password": testPassword,
	}

	jsonData, _ := json.Marshal(loginData)
	req, err := http.NewRequest("POST", userServiceURL+"/user/login", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	return loginResp.Token
}

func createExpense(t *testing.T, token string) string {
	expenseData := map[string]interface{}{
		"name":        "Test Expense",
		"description": "Integration test expense",
		"amount":      50.0,
		"category":    "test",
		"date":        time.Now().UTC().Format("2006-01-02"), // Required by current validation (will be optional after deploy)
		"tags":        []string{"integration-test"},
	}

	jsonData, _ := json.Marshal(expenseData)
	req, _ := http.NewRequest("POST", expensesServiceURL+"/expenses", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create expense: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create expense, status: %d", resp.StatusCode)
	}

	var expense Expense
	if err := json.NewDecoder(resp.Body).Decode(&expense); err != nil {
		t.Fatalf("Failed to decode expense: %v", err)
	}

	return expense.ExpenseId
}
