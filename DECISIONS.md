# Architectural Decision Records (ADR)

## Why These Changes?

### Decision 1: Keep Separate Services (Don't Merge)
**Status**: Accepted

**Context**: 
You have 4 services (user, expense, payment, bank) that could potentially be merged.

**Decision**: Keep them separate

**Reasoning**:
- ✅ Better for learning microservices patterns
- ✅ Each has distinct business responsibility
- ✅ Can scale independently
- ✅ Shows understanding of service boundaries in interviews
- ✅ Allows different teams to own different services (real-world scenario)

**Consequences**:
- More complexity in orchestration
- Need proper service communication patterns
- Worth it for learning and demonstration purposes

---

### Decision 2: Add AI-Analytics Service (New)
**Status**: Accepted

**Context**: 
Want to add LLM/AI capabilities to the project.

**Decision**: Create dedicated AI-Analytics service instead of embedding in Expense service

**Reasoning**:
- ✅ **Separation of Concerns**: Analytics is separate from transactional operations
- ✅ **Resource Management**: AI/LLM calls are expensive, can be scaled separately
- ✅ **Async Processing**: AI analysis can happen in background without blocking main flow
- ✅ **Future Flexibility**: Easy to swap LLM providers or add more AI features
- ✅ **Interview Appeal**: Hot topic, demonstrates modern tech stack
- ✅ **Learning**: RAG, embeddings, prompt engineering, vector databases

**Alternatives Considered**:
- Embed in Expense Service → Would violate single responsibility
- Use external SaaS → Less learning, less impressive

**Implementation**:
- Start with simple categorization
- Gradually add analysis and NL queries
- Can use free Ollama locally for development

---

### Decision 3: Add API Gateway (New)
**Status**: Accepted

**Context**: 
Currently clients call services directly.

**Decision**: Add API Gateway as single entry point

**Reasoning**:
- ✅ **Single Endpoint**: Simplifies client integration
- ✅ **Centralized Auth**: Validate JWT once at gateway
- ✅ **Rate Limiting**: Easier to implement per-user limits
- ✅ **API Versioning**: Can support /v1, /v2 without changing services
- ✅ **Standard Pattern**: Used in all major microservices architectures
- ✅ **Future-Ready**: Easy to add GraphQL, WebSocket, etc.

**Alternatives Considered**:
- Use Istio for routing → Could work, but less control and harder to customize
- Use Kong/Traefik → Good production choice, but building our own shows more coding skill

**Trade-offs**:
- Single point of failure (mitigated with multiple replicas)
- Additional network hop (negligible latency ~5ms)
- More code to maintain (worth it for learning)

---

### Decision 4: Implement Full Middleware Layer
**Status**: Accepted

**Context**: 
Currently minimal middleware, lots of repeated code.

**Decision**: Build comprehensive reusable middleware

**Components**:
1. **Auth Middleware** - JWT validation, user context injection
2. **Rate Limiter** - Prevent abuse, per-user/IP limits
3. **Circuit Breaker** - Fail fast on downstream failures
4. **Request ID** - Track requests across services
5. **Validation** - Input validation before handler
6. **Recovery** - Graceful panic handling

**Reasoning**:
- ✅ **DRY Principle**: Write once, use everywhere
- ✅ **Production-Ready**: These are must-haves in real systems
- ✅ **Interview Topics**: Shows understanding of cross-cutting concerns
- ✅ **Resilience**: Makes system more robust
- ✅ **Observability**: Better logging, tracing, debugging

**Learning Outcomes**:
- Understanding of middleware pattern
- Resilience patterns (circuit breaker, retry, timeout)
- Security best practices
- Performance optimization

---

### Decision 5: Event-Driven Architecture (Optional)
**Status**: Proposed (Optional in Phase 6)

**Context**: 
Currently using synchronous HTTP calls for all communication.

**Decision**: Add message queue for async operations

**When to Use**:
- ✅ Expense Created → AI Analysis (can be async)
- ✅ Payment Processed → Update Expense Status
- ✅ Bulk operations (future)
- ❌ User validation → Keep synchronous (need immediate response)

**Reasoning**:
- ✅ **Loose Coupling**: Services don't need to know about each other
- ✅ **Async Processing**: Don't block user while AI analyzes
- ✅ **Reliability**: Messages are persisted, can retry
- ✅ **Scalability**: Can process events in parallel
- ✅ **Modern Pattern**: Event-driven is trending in microservices

**Why Optional**:
- Adds complexity
- Need to manage message queue infrastructure
- Can be added later without breaking existing code
- Start with HTTP, add events for specific use cases

**Recommendation**: Start with Phase 1-5, add events in Phase 6 if needed.

---

### Decision 6: Remove/Repurpose HydraDB Service
**Status**: Proposed

**Context**: 
HydraDB service exists but unclear purpose.

**Options**:
1. **Remove entirely** if unused
2. **Repurpose as Redis Service** for caching
3. **Use for event store** if doing event sourcing

**Recommendation**: Remove for now, we'll use ElastiCache/Redis for caching

**Reasoning**:
- If it's not providing clear value, remove it
- Simpler architecture is better
- Can always add back if needed

---

### Decision 7: Use Shared HTTP Client Utility
**Status**: Accepted

**Context**: 
Services make HTTP calls with inconsistent patterns.

**Decision**: Create `app/utils/httpclient` package

**Features**:
- Context-aware (timeout, cancellation)
- Automatic retry with exponential backoff
- Circuit breaker integration
- Distributed tracing propagation
- Structured logging

**Reasoning**:
- ✅ **Consistency**: All services use same patterns
- ✅ **Reliability**: Retry and circuit breaker built-in
- ✅ **Observability**: Automatic tracing and logging
- ✅ **Maintainability**: Fix once, benefits all services

**Example**:
```go
// Instead of http.Get(url)
client := httpclient.New(
    httpclient.WithTimeout(5*time.Second),
    httpclient.WithRetries(3),
)
resp, err := client.Get(ctx, url)
```

---

### Decision 8: Use bcrypt for Password Hashing
**Status**: Accepted

**Context**: 
Currently storing passwords in plain text (CRITICAL ISSUE!)

**Decision**: Use bcrypt with appropriate cost factor

**Reasoning**:
- ✅ Industry standard for password hashing
- ✅ Built-in salt generation
- ✅ Configurable work factor
- ✅ Resistant to brute force attacks

**Alternatives Considered**:
- argon2 → More secure but overkill for learning project
- SHA256 → NOT suitable for passwords (too fast)

**Implementation**:
```go
// Cost factor 12 = good balance between security and performance
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(password), 
    bcrypt.DefaultCost, // 10, or use 12 for more security
)
```

---

### Decision 9: JWT Secret in AWS Secrets Manager
**Status**: Accepted

**Context**: 
Currently JWT secret is hardcoded in code.

**Decision**: Store in environment variable (dev) or AWS Secrets Manager (prod)

**Reasoning**:
- ✅ Security best practice
- ✅ Can rotate secrets without code changes
- ✅ Different secrets per environment
- ✅ Audit trail of secret access

**Implementation**:
```go
// Development
jwtSecret := []byte(os.Getenv("JWT_SECRET_KEY"))

// Production
import "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
// Fetch from Secrets Manager
```

---

### Decision 10: Use OpenAI API (or Ollama) for LLM
**Status**: Accepted

**Context**: 
Need LLM integration for AI features.

**Decision**: Start with OpenAI API, support Ollama for local dev

**Options Comparison**:

| Provider | Cost | Quality | Local | Learning Value |
|----------|------|---------|-------|----------------|
| OpenAI GPT-4 | $$$ | Excellent | No | High (industry standard) |
| Anthropic Claude | $$$ | Excellent | No | High (alternative approach) |
| Ollama (Llama 3) | Free | Good | Yes | Medium (local models) |
| AWS Bedrock | $$$ | Good | No | Medium (AWS integration) |

**Decision**:
- **Development**: Use Ollama (free, local)
- **Production**: Use OpenAI API (best quality)
- **Support both** via interface

**Reasoning**:
- ✅ Free local development
- ✅ Best quality for demo/production
- ✅ Learn both approaches
- ✅ Easy to swap providers

**Interface**:
```go
type LLMProvider interface {
    Complete(ctx context.Context, prompt string) (string, error)
    Embed(ctx context.Context, text string) ([]float64, error)
}

// Implementations: OpenAIProvider, OllamaProvider
```

---

### Decision 11: Use Redis for Caching
**Status**: Accepted

**Context**: 
LLM queries are expensive, need caching.

**Decision**: Add Redis for caching (ElastiCache in production)

**Cache Strategy**:
```
Query → Check Cache → Cache Hit? → Return
                    ↓ Cache Miss
                    → Call LLM → Cache Result → Return
```

**TTL Strategy**:
- LLM responses: 1 hour
- User data: 5 minutes
- Expense data: 10 minutes
- Analysis results: 30 minutes

**Reasoning**:
- ✅ Reduce LLM API costs (expensive!)
- ✅ Improve response time
- ✅ Reduce load on services
- ✅ Standard caching solution

---

### Decision 12: Test Strategy
**Status**: Accepted

**Context**: 
Currently no tests in project.

**Decision**: Add comprehensive testing

**Priorities**:
1. **Unit Tests** (70% coverage target)
   - All business logic in services
   - Critical utility functions
   
2. **Integration Tests**
   - Repository layer (DynamoDB interactions)
   - HTTP client calls
   
3. **E2E Tests**
   - Critical user flows
   - Happy path + error cases
   
4. **Load Tests** (optional)
   - Performance benchmarks
   - Scalability validation

**Tools**:
- `testing` package (built-in)
- `testify` for assertions
- `gomock` for mocking
- `httptest` for HTTP testing
- `k6` for load testing

**Reasoning**:
- ✅ Confidence in code changes
- ✅ Documentation of expected behavior
- ✅ Catch regressions early
- ✅ Required for production systems
- ✅ Shows professionalism in interviews

---

## Interview Talking Points

After this refactor, you can discuss:

### Architecture & Design
- "I built a microservices architecture with 6 services..."
- "Implemented API Gateway pattern for centralized auth and routing..."
- "Used event-driven architecture for async processing..."

### AI/LLM Integration
- "Integrated OpenAI GPT-4 for intelligent expense categorization..."
- "Implemented RAG pattern with vector embeddings..."
- "Built natural language query interface..."

### Resilience & Reliability
- "Implemented circuit breaker pattern to handle downstream failures..."
- "Added retry logic with exponential backoff..."
- "Used rate limiting to prevent abuse..."

### Security
- "Implemented JWT authentication with middleware..."
- "Used bcrypt for password hashing..."
- "Stored secrets in AWS Secrets Manager..."

### Observability
- "Used OpenTelemetry for distributed tracing across services..."
- "Implemented structured logging with request correlation..."
- "Set up Prometheus metrics and Grafana dashboards..."

### Cloud & DevOps
- "Deployed on AWS EKS with Istio service mesh..."
- "Used GitOps with FluxCD for automated deployments..."
- "Implemented IaC with Terraform in multi-stage pipeline..."

---

## Next Steps

1. **Review this plan** - Make sure you understand each decision
2. **Ask questions** - Anything unclear?
3. **Choose starting point** - Which phase first?
4. **Set timeline** - How much time per week?

**Recommendation**: Start with Phase 1 (Critical Fixes) this week!

Ready to begin? 🚀





