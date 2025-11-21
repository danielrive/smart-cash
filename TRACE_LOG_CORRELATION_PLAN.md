# Trace-Log Correlation Plan

## Current State Analysis

### ✅ What's Working

1. **HTTP Middleware Logs** (Automatic)
   - Location: `app/utils/logging/middleware.go`
   - Status: ✅ Already includes `trace_id` and `span_id`
   - Example log:
   ```json
   {
     "time": "2024-01-15T10:30:00Z",
     "level": "INFO",
     "msg": "request processed",
     "method": "GET",
     "path": "/users/123",
     "status": 200,
     "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
     "span_id": "00f067aa0ba902b7"
   }
   ```

2. **Helper Functions Available**
   - `GetTraceContext(ctx)` - Extracts trace_id and span_id from context
   - `WithTraceContext(ctx)` - Returns attributes for inline use
   - `LoggerWithTrace(ctx, logger)` - Creates logger with trace context pre-attached

### ❌ What's Missing

1. **Service Layer Logs** (Manual - Not Using Trace Context)
   - Current: `us.logger.Debug("getting user by id", ...)`
   - Missing: `trace_id` and `span_id` in logs
   - Impact: Can't correlate service logs with traces

2. **Handler Layer Logs** (Manual - Not Using Trace Context)
   - Current: `h.logger.Info("getting user by id", ...)`
   - Missing: `trace_id` and `span_id` in logs
   - Impact: Can't correlate handler logs with traces

3. **Repository Layer Logs** (Manual - Not Using Trace Context)
   - Current: `r.logger.Debug("creating expense in database", ...)`
   - Missing: `trace_id` and `span_id` in logs
   - Impact: Can't correlate repository logs with traces

## The Problem

When you search logs by `trace_id` in your log aggregation system (like ELK, CloudWatch, etc.), you can only find:
- ✅ HTTP request logs (from middleware)
- ❌ Service logs (missing trace_id)
- ❌ Handler logs (missing trace_id)
- ❌ Repository logs (missing trace_id)

**Result**: You can't see the full picture of what happened in a request!

## Solution Approaches

### Approach 1: Manual - Use Helper Functions (Current Available)

**How it works:**
```go
// Option A: Use LoggerWithTrace helper
traceLogger := logging.LoggerWithTrace(ctx, us.logger)
traceLogger.Debug("getting user by id",
    slog.String("user_id", userId),
    slog.String("component", "service"),
)

// Option B: Use WithTraceContext inline
us.logger.Debug("getting user by id",
    append([]any{
        slog.String("user_id", userId),
        slog.String("component", "service"),
    }, utils.WithTraceContext(ctx)...)...,
)
```

**Pros:**
- ✅ Explicit control
- ✅ Only adds trace context when needed
- ✅ No performance overhead when not used

**Cons:**
- ❌ Requires changing every log call
- ❌ Easy to forget
- ❌ Lots of code changes across all services

**Effort:** High (need to update ~100+ log calls)

---

### Approach 2: Automatic - Custom slog Handler (Recommended)

**How it works:**
Create a custom slog handler that automatically extracts trace context from context and adds it to all logs.

**Implementation:**
```go
// Custom handler that wraps JSONHandler and adds trace context
type TraceHandler struct {
    handler slog.Handler
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
    // Extract trace context from context
    traceAttrs := utils.GetTraceContext(ctx)
    if traceAttrs != nil {
        for _, attr := range traceAttrs {
            r.AddAttrs(attr)
        }
    }
    return h.handler.Handle(ctx, r)
}
```

**Usage:**
```go
// In service - no changes needed!
us.logger.Debug("getting user by id", ...)  // trace_id automatically added!
```

**Pros:**
- ✅ Automatic - works everywhere
- ✅ No code changes needed in services/handlers/repositories
- ✅ Consistent across all logs
- ✅ Zero maintenance

**Cons:**
- ❌ Requires context to be passed to logger (slog supports this!)
- ❌ Need to update logger initialization

**Effort:** Low (update logger init, no service changes)

---

### Approach 3: Hybrid - Context-Aware Logger Methods

**How it works:**
Add methods to logger that accept context:
```go
logger.DebugWithContext(ctx, "message", attrs...)
logger.InfoWithContext(ctx, "message", attrs...)
```

**Pros:**
- ✅ Explicit when trace context is needed
- ✅ Can choose per log call

**Cons:**
- ❌ Still requires code changes
- ❌ Need to remember to use `*WithContext` methods
- ❌ More methods to maintain

**Effort:** Medium (add methods + update log calls)

---

## Recommended Approach: #2 (Automatic Handler)

### Why?

1. **Zero Code Changes** - Services/handlers/repositories don't need changes
2. **Automatic** - Works everywhere, can't forget
3. **Consistent** - All logs have trace context
4. **Future-proof** - New code automatically gets trace context

### Implementation Plan

#### Step 1: Create Custom Trace Handler
- Location: `app/utils/logging/trace_handler.go`
- Wraps JSONHandler
- Extracts trace context from context automatically

#### Step 2: Update Logger Initialization
- Location: `app/utils/logging/logger.go`
- Use TraceHandler instead of JSONHandler directly
- No changes to InitLogger signature

#### Step 3: Update Log Calls to Use Context (if needed)
- Check if slog supports context in Handle method
- If yes: No changes needed!
- If no: Use `logger.WithContext(ctx).Debug(...)` pattern

#### Step 4: Test
- Verify trace_id appears in all logs
- Test in Jaeger + log aggregation system

### Technical Details

**slog Context Support:**
- slog.Handler.Handle() receives `context.Context` as first parameter
- We can extract trace context from this context
- Perfect for automatic trace correlation!

**Example Handler:**
```go
type TraceHandler struct {
    handler slog.Handler
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
    // Get trace context from context
    if traceAttrs := utils.GetTraceContext(ctx); traceAttrs != nil {
        for _, attr := range traceAttrs {
            r.AddAttrs(attr)
        }
    }
    return h.handler.Handle(ctx, r)
}

func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.handler.Enabled(ctx, level)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &TraceHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *TraceHandler) WithGroup(name string) slog.Handler {
    return &TraceHandler{handler: h.handler.WithGroup(name)}
}
```

## Decision Matrix

| Approach | Effort | Maintenance | Consistency | Recommended |
|----------|--------|-------------|-------------|-------------|
| Manual (Helper) | High | Medium | Low | ❌ |
| Automatic (Handler) | Low | Low | High | ✅ |
| Hybrid (Methods) | Medium | Medium | Medium | ⚠️ |

## Next Steps

1. ✅ Review this plan
2. ⏳ Decide on approach
3. ⏳ Implement chosen approach
4. ⏳ Test in all services
5. ⏳ Verify in Jaeger + logs

## Questions to Consider

1. Do we want automatic trace context in ALL logs, or only specific ones?
2. Should we add trace context even when there's no active span?
3. Do we want to add trace context to logs that don't have context available?

