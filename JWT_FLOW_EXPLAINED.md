# JWT Token Complete Flow - Explained 🔐

## Overview

JWT (JSON Web Token) is used to authenticate users across all services. Here's the complete lifecycle:

---

## 🔄 Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    STEP 1: REGISTRATION                         │
│                     (No Token Yet)                              │
└─────────────────────────────────────────────────────────────────┘

User → POST /user {username, email, password}
              ↓
        User Service
              ↓
        repositories/user.go:CreateUser()
              ↓
        DynamoDB: Save user data
              ↓
        Return: {userId: "abc-123", username: "john"}
        
❌ NO TOKEN! User must login to get a token

┌─────────────────────────────────────────────────────────────────┐
│                    STEP 2: LOGIN                                │
│                 (Token Creation)                                │
└─────────────────────────────────────────────────────────────────┘

User → POST /user/login {username, password}
              ↓
        handler/handler.go:Login()
              ↓
        service/user.go:Login() [line 78]
              ↓
        1. GetUserByEmailorUsername() [line 83]
           → Fetch user from DynamoDB
              ↓
        2. Validate password [line 90]
           ⚠️ Currently plain text (will fix with bcrypt)
           if (stored_password != provided_password) → Error
              ↓
        3. Generate JWT token [line 97]
           us.generateJWT(userId, username, email)
              ↓
           ┌──────────────────────────────────────┐
           │   generateJWT() [line 116-131]      │
           ├──────────────────────────────────────┤
           │  1. Create claims:                   │
           │     - userId: "abc-123"              │
           │     - username: "john"               │
           │     - email: "john@example.com"      │
           │     - exp: now + 1 hour              │
           │                                      │
           │  2. Create token:                    │
           │     jwt.NewWithClaims(HS256, claims) │
           │                                      │
           │  3. Sign with secret:                │
           │     token.SignedString(us.jwtSecret) │
           │                      ▲                │
           │                      │                │
           │             USES JWT_SECRET!         │
           │         (from environment var)       │
           └──────────────────────────────────────┘
              ↓
        Return: {token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

┌─────────────────────────────────────────────────────────────────┐
│                  STEP 3: USING TOKEN                            │
│               (Token Validation)                                │
└─────────────────────────────────────────────────────────────────┘

User → GET /expenses
       Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
              ↓
        Expenses Service
              ↓
        middleware.AuthMiddleware() runs FIRST
              ↓
           ┌──────────────────────────────────────┐
           │   AuthMiddleware Validation          │
           ├──────────────────────────────────────┤
           │  1. Extract header [line 29]         │
           │     Authorization: Bearer <token>    │
           │                                      │
           │  2. Extract token [line 39]          │
           │     Remove "Bearer " prefix          │
           │                                      │
           │  3. Parse & Validate [line 49]       │
           │     jwt.ParseWithClaims(token...)    │
           │                      ▼                │
           │     Validates:                       │
           │     - Signature (using jwtSecret)    │
           │     - Algorithm (must be HS256)      │
           │     - Expiration (must be < exp)     │
           │     - Structure (valid JSON)         │
           │                                      │
           │  4. If valid [line 77]:              │
           │     c.Set("userId", claims.UserID)   │
           │     c.Set("username", claims.Username)│
           │     c.Set("email", claims.Email)     │
           │                                      │
           │  5. Continue [line 82]               │
           │     c.Next() → Call handler          │
           └──────────────────────────────────────┘
              ↓
        expensesHandler.CreateExpense()
              ↓
        Can now access:
        userId := c.GetString("userId")    // "abc-123"
        username := c.GetString("username") // "john"
              ↓
        Create expense for this user
              ↓
        Return: {expenseId: "...", userId: "abc-123", ...}
```

---

## 🔑 The Secret Key - How It Works

### **Token Creation & Validation MUST Use Same Secret**

```
┌─────────────────────────────────────────────────────────────────┐
│                      THE SECRET KEY                             │
└─────────────────────────────────────────────────────────────────┘

Environment Variable: JWT_SECRET
                      ▲
                      │
        ┌─────────────┴─────────────┐
        │                           │
   main.go                     main.go
   (user-service)          (expenses-service)
        │                           │
        ├── Loads JWT_SECRET        ├── Loads JWT_SECRET
        ├── Passes to service       ├── Passes to middleware
        │                           │
        ▼                           ▼
   service/user.go           middleware/auth.go
        │                           │
        ├── generateJWT()           ├── AuthMiddleware()
        │   Signs token              │   Validates signature
        │   with secret              │   with secret
        │                           │
        ▼                           ▼
   Token: "eyJ..."             Valid? ✅ YES
                                      (secrets match!)
```

### **What Happens If Secrets Don't Match?**

```
❌ SCENARIO: Mismatched Secrets

User Service:     JWT_SECRET = "secret-123"
Expenses Service: JWT_SECRET = "different-456"

Login:
  → Generates token signed with "secret-123"
  → Returns token to user

Request:
  → User sends token to expenses-service
  → Middleware tries to validate with "different-456"
  → Signature doesn't match!
  → Returns 401 Unauthorized

🔴 TOKEN ALWAYS REJECTED!
```

---

## 📝 Token Anatomy

### **What's Inside a JWT Token?**

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWJjLTEyMyIsInVzZXJuYW1lIjoiam9obiIsImVtYWlsIjoiam9obkBleGFtcGxlLmNvbSIsImV4cCI6MTY5OTk5OTk5OX0.signature_here
└──────────── HEADER ──────────────┘ └────────────────────────── PAYLOAD ─────────────────────────────┘ └── SIGNATURE ──┘
```

### **1. Header** (Base64 encoded)
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### **2. Payload (Claims)** (Base64 encoded)
```json
{
  "user_id": "abc-123",
  "username": "john",
  "email": "john@example.com",
  "exp": 1699999999
}
```

### **3. Signature**
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  JWT_SECRET
)
```

**The signature proves:**
- Token was created by someone who knows the secret
- Token hasn't been tampered with
- If you change payload, signature becomes invalid

---

## 🔐 Security Features

### **What We Validate:**

1. ✅ **Signature Verification**
   - Token signed with correct secret?
   - If secret is wrong → Invalid signature

2. ✅ **Algorithm Verification**
   - Must be HS256
   - Prevents "algorithm confusion" attacks

3. ✅ **Expiration Check**
   - Token must not be expired
   - We set: 1 hour from creation
   - After 1 hour → Must login again

4. ✅ **Structure Validation**
   - Valid JSON?
   - Has required fields?
   - Correct format?

---

## 🛡️ What Middleware Does vs What Service Does

### **User Service (Token CREATION)**
**File**: `user-service/internal/service/user.go`

**Responsibilities:**
1. ✅ Validate username/password
2. ✅ Create JWT claims (userId, username, email, exp)
3. ✅ Sign token with JWT secret
4. ✅ Return token to user

**Code Location:**
- `Login()` method - Line 78-111
- `generateJWT()` method - Line 116-131

---

### **Middleware (Token VALIDATION)**
**File**: `app/utils/middleware/auth.go`

**Responsibilities:**
1. ✅ Extract token from Authorization header
2. ✅ Validate token signature
3. ✅ Check token not expired
4. ✅ Extract claims from token
5. ✅ Inject user info into request context

**Code Location:**
- `AuthMiddleware()` function - Line 26-84

---

## 🔍 How to Debug JWT Issues

### **Issue: "missing authorization header"**
```bash
# Check request has header
curl http://localhost:8282/expenses \
  -v  # ← Shows all headers

# Look for:
> Authorization: Bearer eyJ...
```

### **Issue: "invalid or expired token"**

**Check 1: Token expired?**
```bash
# Decode token (online: jwt.io)
# Check "exp" claim
# If exp < current time → Get new token
```

**Check 2: Secrets match?**
```bash
# User service
echo $JWT_SECRET  # Should output secret

# Expenses service  
echo $JWT_SECRET  # Should be SAME

# If different → That's the problem!
```

**Check 3: Token structure?**
```bash
# Token should have 3 parts separated by dots
eyJhbG...  .  eyJ1c2...  .  SflKxw...
└─ Header ┘  └─ Payload ┘  └─ Signature ┘
```

---

## 📊 Files Involved

### **Token Creation:**
- `user-service/main.go` - Loads JWT_SECRET, creates service
- `user-service/internal/service/user.go` - Generates tokens
- `user-service/internal/handler/handler.go` - Login endpoint

### **Token Validation:**
- `app/utils/middleware/auth.go` - Validates tokens
- `expenses-service/main.go` - Uses middleware
- `payment-service/main.go` - Uses middleware
- `bank-service/main.go` - Uses middleware

### **Environment:**
- `JWT_SECRET` environment variable (MUST be same across all services!)

---

## ✅ Summary

**Token Lifecycle:**
1. User registers → Stored in DB (no token)
2. User logs in → Token created & signed with JWT_SECRET
3. User makes request → Token validated with same JWT_SECRET
4. If valid → Request proceeds with user context
5. After 1 hour → Token expires, user must login again

**Key Points:**
- ✅ Token created during login
- ✅ Token signed with secret key
- ✅ Same secret used to validate
- ✅ Token carries user info (userId, username, email)
- ✅ Middleware extracts user info and injects into context
- ✅ Handlers can access user info via `c.GetString("userId")`

**The Fix We Just Made:**
- ❌ Before: User service used hardcoded "123456"
- ✅ After: User service uses JWT_SECRET from environment
- ✅ Now: Creation and validation use SAME secret! 🎉





