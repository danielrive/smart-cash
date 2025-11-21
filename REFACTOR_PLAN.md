# Refactor Plan - Smart Cash Project

## Overview
This document outlines the phased approach to refactoring and enhancing the Smart Cash microservices application.

---

## Phase 1: Critical Fixes ✅ COMPLETED
**Status**: Done
**See**: `PHASE1_COMPLETED.md`

### Completed Items
- ✅ Merge conflict resolution
- ✅ Duplicate middleware removal
- ✅ Error handling fixes
- ✅ Context propagation fixes
- ✅ Type corrections

---

## Phase 2: Middleware & Security ✅ COMPLETED
**Status**: Done
**See**: `AUTH_MIDDLEWARE_APPLIED.md`, `VALIDATION_MIDDLEWARE_APPLIED.md`

### Completed Items
- ✅ JWT Authentication Middleware
- ✅ Request Validation Middleware
- ✅ Path Parameter Validation
- ✅ Query Parameter Validation
- ✅ Duplicate User Prevention

---

## Phase 3: Logging & Observability
**Status**: In Progress

### Objectives
- Implement structured logging across all services
- Add request correlation IDs
- Enhance error logging with context
- Integrate with logging aggregation (CloudWatch/ELK)
- Add performance metrics logging

### Tasks

#### 3.1 Structured Logging
- [ ] Standardize log format across all services
- [ ] Add log levels (DEBUG, INFO, WARN, ERROR)
- [ ] Include request context in all logs
- [ ] Add user context (userId, username) to logs
- [ ] Log all HTTP requests/responses
- [ ] Log all database operations
- [ ] Log all external service calls

#### 3.2 Request Correlation
- [ ] Generate unique request ID per request
- [ ] Propagate request ID across service calls
- [ ] Include request ID in all log entries
- [ ] Add request ID to HTTP headers
- [ ] Display request ID in error responses

#### 3.3 Error Logging Enhancement
- [ ] Add stack traces for errors
- [ ] Include full context in error logs
- [ ] Log error recovery actions
- [ ] Add error categorization (validation, auth, internal, external)
- [ ] Track error rates per endpoint

#### 3.4 Performance Logging
- [ ] Log request duration
- [ ] Log database query times
- [ ] Log external API call durations
- [ ] Add slow query detection
- [ ] Log memory usage for expensive operations

#### 3.5 Log Aggregation
- [ ] Configure CloudWatch Logs (AWS)
- [ ] Set up log groups per service
- [ ] Add log retention policies
- [ ] Create log search/filter capabilities
- [ ] Set up log-based alerts

### Implementation Details

**Log Format Example:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "user-service",
  "request_id": "abc123",
  "user_id": "7cc8bd9a-0a20-4062-af5b-fdc4b08e36ac",
  "message": "User created successfully",
  "duration_ms": 45,
  "trace_id": "trace-xyz"
}
```

**Files to Modify:**
- `app/utils/logger/` (new package)
- All service `main.go` files
- All handler files
- All service files
- All repository files

---

## Phase 4: Concurrency & Performance
**Status**: Planned

### Objectives
- Implement proper concurrency patterns
- Add connection pooling
- Optimize database queries
- Add caching layer
- Implement async processing where appropriate
- Add rate limiting and throttling

### Tasks

#### 4.1 Connection Pooling
- [ ] Configure DynamoDB client connection pooling
- [ ] Set appropriate pool sizes per service
- [ ] Add connection timeout configuration
- [ ] Monitor connection pool metrics
- [ ] Implement connection health checks

#### 4.2 Database Query Optimization
- [ ] Review and optimize DynamoDB queries
- [ ] Add query result caching
- [ ] Implement batch operations where possible
- [ ] Add query pagination for large results
- [ ] Optimize GSI usage

#### 4.3 Concurrent Request Handling
- [ ] Review goroutine usage
- [ ] Add worker pools for background tasks
- [ ] Implement proper context cancellation
- [ ] Add request timeout handling
- [ ] Prevent goroutine leaks

#### 4.4 Caching Strategy
- [ ] Add Redis/ElastiCache integration
- [ ] Cache user data (5 min TTL)
- [ ] Cache expense queries (10 min TTL)
- [ ] Cache LLM responses (1 hour TTL)
- [ ] Implement cache invalidation
- [ ] Add cache hit/miss metrics

#### 4.5 Async Processing
- [ ] Identify operations suitable for async
- [ ] Implement message queue (SQS/RabbitMQ)
- [ ] Add async expense processing
- [ ] Add async AI analysis
- [ ] Implement event publishing
- [ ] Add async job status tracking

#### 4.6 Rate Limiting & Throttling
- [ ] Implement per-user rate limits
- [ ] Add per-IP rate limits
- [ ] Add endpoint-specific limits
- [ ] Implement token bucket algorithm
- [ ] Add rate limit headers to responses
- [ ] Log rate limit violations

#### 4.7 Performance Monitoring
- [ ] Add response time metrics
- [ ] Track throughput per endpoint
- [ ] Monitor database query performance
- [ ] Track cache hit rates
- [ ] Add p95/p99 latency tracking
- [ ] Set up performance alerts

### Implementation Details

**Worker Pool Example:**
```go
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
}

func (wp *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < wp.workers; i++ {
        go wp.worker(ctx)
    }
}
```

**Connection Pool Configuration:**
```go
cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRegion(region),
    config.WithHTTPClient(&http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }),
)
```

**Files to Modify:**
- `app/utils/pool/` (new package)
- `app/utils/cache/` (new package)
- All service `main.go` files
- All repository files
- All handler files

---

## Phase 5: API Gateway
**Status**: Planned

### Objectives
- Create unified API Gateway
- Centralize authentication
- Implement request routing
- Add API versioning
- Centralize rate limiting

### Tasks
- [ ] Design API Gateway architecture
- [ ] Implement routing logic
- [ ] Add service discovery
- [ ] Implement load balancing
- [ ] Add API versioning (/v1, /v2)
- [ ] Centralize JWT validation
- [ ] Add request/response transformation

---

## Phase 6: AI/Analytics Service
**Status**: Planned

### Objectives
- Add intelligent expense categorization
- Implement natural language queries
- Add expense analysis and insights
- Create recommendation engine

### Tasks
- [ ] Design AI service architecture
- [ ] Integrate OpenAI/Ollama
- [ ] Implement expense categorization
- [ ] Add NL query interface
- [ ] Create analysis endpoints
- [ ] Add vector embeddings
- [ ] Implement RAG pattern

---

## Phase 7: Testing
**Status**: Planned

### Objectives
- Achieve 70% code coverage
- Add integration tests
- Add E2E tests
- Add load tests

### Tasks
- [ ] Set up testing framework
- [ ] Write unit tests for services
- [ ] Write repository tests
- [ ] Write handler tests
- [ ] Add integration tests
- [ ] Add E2E test scenarios
- [ ] Set up CI/CD test pipeline

---

## Phase 8: Documentation
**Status**: Planned

### Objectives
- Complete API documentation
- Add architecture diagrams
- Create deployment guides
- Add developer onboarding docs

### Tasks
- [ ] Generate OpenAPI/Swagger docs
- [ ] Create architecture diagrams
- [ ] Write deployment guide
- [ ] Add troubleshooting guide
- [ ] Create runbook for operations

---

## Current Priority: Phase 3 & 4

### Next Steps
1. **Start with Logging (Phase 3)**
   - Implement structured logging
   - Add request correlation
   - Enhance error logging

2. **Then Concurrency (Phase 4)**
   - Add connection pooling
   - Implement caching
   - Optimize queries

### Timeline Estimate
- **Phase 3 (Logging)**: 1-2 weeks
- **Phase 4 (Concurrency)**: 2-3 weeks

---

## Notes

### Logging Best Practices
- Always include context (request_id, user_id, service)
- Use appropriate log levels
- Don't log sensitive data (passwords, tokens)
- Use structured logging (JSON)
- Include timing information

### Concurrency Best Practices
- Always use context for cancellation
- Prevent goroutine leaks
- Use worker pools for background tasks
- Monitor resource usage
- Test under load

---

## References
- See `DECISIONS.md` for architectural decisions
- See `PHASE1_COMPLETED.md` for completed work
- See `AUTH_MIDDLEWARE_APPLIED.md` for auth implementation
- See `VALIDATION_MIDDLEWARE_APPLIED.md` for validation implementation
