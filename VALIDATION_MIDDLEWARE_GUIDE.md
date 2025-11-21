# 🛡️ Validation Middleware Guide

## Overview

The validation middleware provides **type-safe**, **declarative validation** for:
- ✅ Request bodies (JSON)
- ✅ Path parameters (`:userId`, `:expenseId`, etc.)
- ✅ Query parameters (`?email=...&username=...`)
- ✅ HTTP headers

## 🎯 Benefits

### **Before (Manual Validation):**
```go
func (h *Handler) CreateExpense(c *gin.Context) {
    var expense dto.CreateExpenseRequest
    
    // Manual JSON binding
    if err := c.ShouldBindJSON(&expense); err != nil {
        c.JSON(400, gin.H{"error": "bad request"})
        return
    }
    
    // Manual validation
    if expense.Name == "" {
        c.JSON(400, gin.H{"error": "name is required"})
        return
    }
    if expense.Amount <= 0 {
        c.JSON(400, gin.H{"error": "amount must be positive"})
        return
    }
    // ... more manual checks
    
    // Finally, business logic
    result, err := h.service.Create(ctx, expense)
}
```

**Problems:**
- ❌ Repetitive validation code
- ❌ Inconsistent error messages
- ❌ Business logic mixed with validation
- ❌ Hard to maintain

### **After (Validation Middleware):**
```go
// In main.go
router.POST("/expenses", 
    middleware.AuthMiddleware(jwtSecret),
    middleware.ValidateBody[dto.CreateExpenseRequest](), // ← Validation here
    handler.CreateExpense)

// In handler
func (h *Handler) CreateExpense(c *gin.Context) {
    // Get pre-validated body from context
    validatedBody := c.MustGet("validatedBody").(dto.CreateExpenseRequest)
    userId := c.GetString("userId") // From auth middleware
    
    validatedBody.UserId = userId
    
    // Clean business logic only!
    result, err := h.service.Create(ctx, validatedBody)
}
```

**Benefits:**
- ✅ Declarative validation (in DTO struct tags)
- ✅ Consistent error messages
- ✅ Separation of concerns
- ✅ Type-safe with generics
- ✅ Reusable across all services

---

## 📚 Middleware Functions

### 1. **ValidateBody[T any]()** - Request Body Validation

Validates JSON request body against struct `validate` tags.

**Usage:**
```go
router.POST("/expenses", 
    middleware.ValidateBody[dto.CreateExpenseRequest](),
    handler.CreateExpense)
```

**DTO Example:**
```go
type CreateExpenseRequest struct {
    Name        string   `json:"name" validate:"required,max=100"`
    Amount      float64  `json:"amount" validate:"required,gt=0"`
    Description string   `json:"description" validate:"required,max=500"`
    Category    string   `json:"category" validate:"required"`
    Date        string   `json:"date" validate:"required,datetime=2006-01-02"`
    Tags        []string `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
}
```

**Validation Tags:**
- `required` - Field must not be empty/zero
- `max=100` - String max length
- `min=5` - String min length
- `gt=0` - Greater than (for numbers)
- `gte=0` - Greater than or equal
- `lt=100` - Less than
- `lte=100` - Less than or equal
- `email` - Valid email format
- `uuid` - Valid UUID format
- `datetime=2006-01-02` - Valid datetime in specified format
- `oneof=pending paid cancelled` - Value must be one of the list
- `dive` - Validate each element in slice/array

**Error Response:**
```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "Amount",
      "message": "'Amount' must be greater than 0",
      "tag": "gt",
      "value": "-10"
    },
    {
      "field": "Name",
      "message": "'Name' is required",
      "tag": "required",
      "value": ""
    }
  ]
}
```

**Access Validated Body in Handler:**
```go
func (h *Handler) CreateExpense(c *gin.Context) {
    validatedBody, exists := c.Get("validatedBody")
    if !exists {
        c.JSON(400, gin.H{"error": "validation failed"})
        return
    }
    
    expense := validatedBody.(dto.CreateExpenseRequest)
    // Use expense...
}
```

---

### 2. **ValidatePathParam(name, type)** - Path Parameter Validation

Validates path parameters like `:userId`, `:expenseId`.

**Usage:**
```go
router.GET("/expenses/:expenseId", 
    middleware.ValidatePathParam("expenseId", "uuid"),
    handler.GetExpenseById)
```

**Supported Types:**
- `"uuid"` - Must be a valid UUID
- `"alphanum"` - Letters and numbers only
- `"numeric"` - Numbers only
- `"required"` - Not empty (basic check)
- Custom validator tags (e.g., `"email"`, `"min=5"`)

**Error Response:**
```json
{
  "error": "Invalid path parameter: expenseId",
  "details": "'expenseId' must be a valid uuid",
  "value": "invalid-id-123"
}
```

**Example:**
```go
// User ID must be UUID
router.GET("/user/:userId", 
    middleware.ValidatePathParam("userId", "uuid"),
    handler.GetUser)

// Transaction ID must be alphanumeric
router.GET("/transaction/:txId", 
    middleware.ValidatePathParam("txId", "alphanum"),
    handler.GetTransaction)
```

---

### 3. **ValidateQueryParams(rules)** - Query Parameter Validation

Validates query parameters with custom rules.

**Usage:**
```go
router.GET("/expenses", 
    middleware.ValidateQueryParams(map[string]string{
        "page":   "required,numeric",
        "limit":  "required,numeric,min=1,max=100",
        "sortBy": "omitempty,oneof=date amount category",
    }),
    handler.ListExpenses)
```

**Example Request:**
```
GET /expenses?page=1&limit=50&sortBy=date
```

**Error Response:**
```json
{
  "error": "Query parameter validation failed",
  "details": [
    {
      "field": "limit",
      "message": "'limit' must be a valid numeric",
      "tag": "numeric",
      "value": "abc"
    }
  ]
}
```

---

### 4. **RequireOneOfQueryParams(params)** - At Least One Query Param

Ensures at least one of the specified query params is present.

**Usage:**
```go
// User must search by email OR username
router.GET("/user", 
    middleware.RequireOneOfQueryParams([]string{"email", "username"}),
    handler.GetUserByQuery)

// Expense must filter by userId OR category
router.GET("/expenses", 
    middleware.RequireOneOfQueryParams([]string{"userId", "category"}),
    handler.GetExpensesByQuery)
```

**Valid Requests:**
```
GET /user?email=john@example.com       ✅
GET /user?username=john                ✅
GET /user?email=john@example.com&username=john ✅
GET /user                              ❌ (no query params)
```

**Error Response:**
```json
{
  "error": "Missing required query parameter",
  "details": "At least one of the following query parameters is required: email, username"
}
```

---

### 5. **ValidateHeader(name, type)** - HTTP Header Validation

Validates HTTP headers.

**Usage:**
```go
router.POST("/expense", 
    middleware.ValidateHeader("X-Request-ID", "uuid"),
    middleware.ValidateHeader("X-API-Version", "required"),
    handler.CreateExpense)
```

**Error Response:**
```json
{
  "error": "Invalid header value: X-Request-ID",
  "details": "'X-Request-ID' must be a valid uuid",
  "value": "invalid-uuid"
}
```

---

## 🔧 Integration Examples

### **Expenses Service (Complete Example)**

```go
// main.go
import (
    "smart-cash/expenses-service/internal/handler"
    "smart-cash/expenses-service/internal/handler/dto"
    "smart-cash/utils/middleware"
)

func main() {
    router := gin.New()
    
    // Health check (no validation)
    router.GET("/expenses/health", handler.HealthCheck)
    
    // Create expense (validate body)
    router.POST("/expenses", 
        middleware.AuthMiddleware(jwtSecret),
        middleware.ValidateBody[dto.CreateExpenseRequest](),
        handler.CreateExpense)
    
    // Get by ID (validate path param)
    router.GET("/expenses/:expenseId", 
        middleware.AuthMiddleware(jwtSecret),
        middleware.ValidatePathParam("expenseId", "uuid"),
        handler.GetExpenseById)
    
    // Query expenses (validate at least one query param)
    router.GET("/expenses", 
        middleware.AuthMiddleware(jwtSecret),
        middleware.RequireOneOfQueryParams([]string{"userId", "category"}),
        handler.GetExpensesByQuery)
    
    // Delete (validate path param)
    router.DELETE("/expenses/:expenseId", 
        middleware.AuthMiddleware(jwtSecret),
        middleware.ValidatePathParam("expenseId", "uuid"),
        handler.DeleteExpense)
    
    router.Run(":8282")
}
```

```go
// internal/handler/dto/expense.go
package dto

type CreateExpenseRequest struct {
    UserId      string   `json:"-"` // From auth middleware
    Name        string   `json:"name" validate:"required,max=100"`
    Amount      float64  `json:"amount" validate:"required,gt=0"`
    Description string   `json:"description" validate:"required,max=500"`
    Category    string   `json:"category" validate:"required"`
    Date        string   `json:"date" validate:"required,datetime=2006-01-02"`
    Tags        []string `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
}
```

```go
// internal/handler/expense.go
func (h *ExpensesHandler) CreateExpense(c *gin.Context) {
    // Get userId from auth middleware
    userId := c.GetString("userId")
    
    // Get pre-validated body
    validatedBody := c.MustGet("validatedBody").(dto.CreateExpenseRequest)
    validatedBody.UserId = userId
    
    // Business logic only
    response, err := h.expensesService.CreateExpense(ctx, validatedBody)
    if err != nil {
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }
    
    c.JSON(201, response)
}
```

---

### **User Service Example**

```go
// main.go
router.POST("/user",
    middleware.ValidateBody[models.User](),
    handler.CreateUser)

router.GET("/user/:userId", 
    middleware.AuthMiddleware(jwtSecret),
    middleware.ValidatePathParam("userId", "uuid"),
    handler.GetUserById)

router.GET("/user", 
    middleware.AuthMiddleware(jwtSecret),
    middleware.RequireOneOfQueryParams([]string{"email", "username"}),
    handler.GetUserByQuery)

router.POST("/login",
    middleware.ValidateBody[models.LoginRequest](),
    handler.Login)
```

```go
// models/user.go
type User struct {
    Username  string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8,max=100"`
    FirstName string `json:"firstName" validate:"required,max=50"`
    LastName  string `json:"lastName" validate:"required,max=50"`
}

type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}
```

---

### **Payment Service Example**

```go
// main.go
router.POST("/payment",
    middleware.AuthMiddleware(jwtSecret),
    middleware.ValidateBody[models.PaymentRequest](),
    handler.ProcessPayment)

router.GET("/payment/:transactionId", 
    middleware.AuthMiddleware(jwtSecret),
    middleware.ValidatePathParam("transactionId", "uuid"),
    handler.GetTransaction)
```

```go
// models/payment.go
type PaymentRequest struct {
    UserId    string  `json:"userId" validate:"required,uuid"`
    ExpenseId string  `json:"expenseId" validate:"required,uuid"`
    Amount    float64 `json:"amount" validate:"required,gt=0"`
    Method    string  `json:"method" validate:"required,oneof=card bank cash"`
}
```

---

## 🎨 Advanced Patterns

### **Combining Multiple Validations**

```go
router.POST("/expense/:expenseId/comment",
    middleware.AuthMiddleware(jwtSecret),
    middleware.ValidatePathParam("expenseId", "uuid"),
    middleware.ValidateBody[dto.CommentRequest](),
    middleware.ValidateHeader("X-Request-ID", "uuid"),
    handler.AddComment)
```

### **Conditional Validation**

```go
// If you need complex validation, do it in handler after basic validation
func (h *Handler) UpdateExpense(c *gin.Context) {
    expense := c.MustGet("validatedBody").(dto.UpdateExpenseRequest)
    expenseId := c.Param("expenseId")
    userId := c.GetString("userId")
    
    // Complex business rule validation
    existingExpense, _ := h.service.GetById(ctx, expenseId)
    if existingExpense.UserId != userId {
        c.JSON(403, gin.H{"error": "forbidden"})
        return
    }
    
    // Continue...
}
```

### **Custom Validation Functions**

If you need custom validation logic not covered by tags, add it to the validator:

```go
// middleware/validation.go
func init() {
    validate = validator.New()
    
    // Register custom validation
    validate.RegisterValidation("customtag", func(fl validator.FieldLevel) bool {
        // Custom logic
        return fl.Field().String() != "forbidden"
    })
}
```

---

## ✅ Migration Checklist

When migrating existing handlers to use validation middleware:

1. **Create/Update DTOs** with `validate` tags
   ```go
   type CreateRequest struct {
       Name string `json:"name" validate:"required,max=100"`
   }
   ```

2. **Add middleware to route**
   ```go
   router.POST("/resource", 
       middleware.ValidateBody[dto.CreateRequest](),
       handler.Create)
   ```

3. **Update handler to use validated body**
   ```go
   func (h *Handler) Create(c *gin.Context) {
       body := c.MustGet("validatedBody").(dto.CreateRequest)
       // Use body...
   }
   ```

4. **Remove manual validation** from handler
   ```go
   // ❌ DELETE THIS
   if err := c.ShouldBindJSON(&body); err != nil { ... }
   if body.Name == "" { ... }
   ```

5. **Test validation errors**
   ```bash
   # Missing required field
   curl -X POST http://localhost:8282/expenses \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"amount": 100}'
   
   # Should return clear validation error
   ```

---

## 📊 Validation Error Format

All validation errors follow this consistent format:

```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "Amount",
      "message": "'Amount' must be greater than 0",
      "tag": "gt",
      "value": "-10"
    }
  ]
}
```

This makes it easy for frontend to:
- Display field-specific errors
- Highlight invalid fields
- Show user-friendly messages

---

## 🧪 Testing Validation

```go
// Example test
func TestCreateExpense_InvalidAmount(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    body := `{"name":"Test","amount":-10}`
    c.Request = httptest.NewRequest("POST", "/expenses", strings.NewReader(body))
    
    // Apply validation middleware
    middleware.ValidateBody[dto.CreateExpenseRequest]()(c)
    
    // Should abort with 400
    assert.Equal(t, 400, w.Code)
    
    // Check error message
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "Validation failed", response["error"])
}
```

---

## 🎯 Best Practices

1. **Always validate at the edge** (middleware layer)
2. **Use DTOs for requests**, not domain models
3. **Keep validation rules in struct tags** for visibility
4. **Add business logic validation in handlers/services**
5. **Return consistent error formats**
6. **Log validation failures** for monitoring
7. **Test validation rules** with unit tests

---

## 📝 Summary

**What We Built:**
- ✅ Type-safe validation middleware using Go generics
- ✅ Declarative validation with struct tags
- ✅ Consistent error messages
- ✅ Path, query, header, and body validation
- ✅ Reusable across all microservices

**What You Learned:**
- ✅ Middleware composition patterns
- ✅ Go generics for type safety
- ✅ Validator library (`go-playground/validator`)
- ✅ Separation of concerns (validation vs business logic)
- ✅ Clean architecture patterns

**Next Steps:**
- Apply to all services (user, payment, bank)
- Add more custom validation rules as needed
- Integrate with frontend error handling
- Add request ID middleware (Phase 2 next item)





