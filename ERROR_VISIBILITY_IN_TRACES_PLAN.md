# Error Visibility in Traces - Plan

## Current Problem

When you look at a trace in Jaeger:

**Success Case (What you see now):**
- ✅ Logs section shows events clearly:
  - `user.id.generated` event
  - `user created successfully` event
  - All with nice attributes visible

**Error Case (What's missing):**
- ❌ Errors might not show up clearly in Logs section
- ❌ Hard to quickly identify which span failed
- ❌ Error details might be buried
- ❌ No clear error event like we have success events

## Current Implementation

### What We Have Now

```go
// In services/repositories when error occurs:
if err != nil {
    utils.RecordSpanError(ctx, err)
    // ... log error
    return err
}
```

### What `RecordSpanError` Currently Does

```go
func RecordSpanError(ctx context.Context, err error) {
    span.RecordError(err)              // Records exception event (with stack trace)
    span.SetStatus(codes.Error, ...)   // Sets span status to Error
}
```

**What this gives us:**
- ✅ Exception event recorded (with stack trace)
- ✅ Span status set to Error
- ✅ Error visible in Jaeger

**What's missing:**
- ❌ Exception event might not be as visible as success events in Logs section
- ❌ No clear error event like "operation failed" (similar to "user created successfully")
- ❌ No error classification/type
- ❌ Limited error context (just the error message)
- ❌ Can't easily see error in Logs timeline like success events

## What We Want in Jaeger

### Visual Indicators
1. **Red/Error Status** - Span should show as error (✅ we have this)
2. **Error Badge/Icon** - Visual indicator in span list
3. **Error Details** - Click span to see error details
4. **Error Type** - Classification (validation, not_found, internal, etc.)
5. **Error Context** - What operation failed, component, etc.

### Information We Want to See
- Error message
- Error type/classification
- Component where error occurred (service, repository, handler)
- Operation that failed
- Any relevant context (user_id, expense_id, etc.)

## Proposed Solutions

### Option 1: Enhance `RecordSpanError` (Recommended)

**Add more attributes and events to make errors more visible:**

```go
func RecordSpanError(ctx context.Context, err error) {
    span := trace.SpanFromContext(ctx)
    if !span.IsRecording() {
        return
    }

    // 1. Record error (adds exception event with stack trace)
    span.RecordError(err)

    // 2. Set error status
    span.SetStatus(codes.Error, err.Error())

    // 3. Add error attributes for filtering/searching
    span.SetAttributes(
        attribute.Bool("error", true),                    // Easy filtering
        attribute.String("error.message", err.Error()),   // Error message
        attribute.String("error.type", classifyError(err)), // Error type
    )

    // 4. Add clear error event (similar to success events like "user created successfully")
    span.AddEvent("operation.failed",
        trace.WithAttributes(
            attribute.String("error.message", err.Error()),
            attribute.String("error.type", classifyError(err)),
        ),
    )
    
    // Note: span.RecordError() already adds exception event with stack trace
    // This AddEvent adds a clear, visible event in Logs section
}
```

**Benefits:**
- ✅ More visible in Jaeger (error attribute, event)
- ✅ Can filter traces by error type
- ✅ Better error classification
- ✅ Minimal code changes (just enhance one function)

---

### Option 2: Add Context to Error Recording

**Pass additional context when recording errors:**

```go
// New function signature
func RecordSpanErrorWithContext(ctx context.Context, err error, component string, operation string, attrs ...attribute.KeyValue)

// Usage:
utils.RecordSpanErrorWithContext(ctx, err, "repository", "GetUserById",
    attribute.String("user.id", userId),
)
```

**Benefits:**
- ✅ More context about where/why error occurred
- ✅ Better debugging information

**Cons:**
- ❌ More verbose (need to pass component/operation)
- ❌ Need to update all error recording calls

---

### Option 3: Automatic Error Context Extraction

**Extract context from span name and attributes:**

```go
func RecordSpanError(ctx context.Context, err error) {
    span := trace.SpanFromContext(ctx)
    
    // Extract component from span attributes
    component := extractComponent(span)
    operation := extractOperation(span)
    
    // Add error with context
    span.SetAttributes(
        attribute.Bool("error", true),
        attribute.String("error.component", component),
        attribute.String("error.operation", operation),
        // ... more attributes
    )
}
```

**Benefits:**
- ✅ Automatic context extraction
- ✅ No code changes needed in services/repositories

**Cons:**
- ❌ Might not always extract correct context
- ❌ More complex implementation

---

## Recommended Approach: Option 1 (Enhanced RecordSpanError)

### Why?

1. **Simple** - Just enhance one function
2. **Effective** - Makes errors much more visible
3. **No Breaking Changes** - Existing code continues to work
4. **Better Visibility** - Errors stand out in Jaeger

### What to Add

1. **Error Attributes:**
   - `error: true` - Boolean flag for easy filtering
   - `error.message` - The error message
   - `error.type` - Classification (validation_error, not_found, internal_error, etc.)

2. **Error Event:**
   - `error.occurred` event with error details
   - Makes error visible in span timeline

3. **Error Classification:**
   - Categorize errors by type
   - Helps with filtering and analysis

### Error Classification Examples

```go
// Classify errors by type
- "validation_error" - Input validation failed
- "not_found" - Resource not found
- "auth_error" - Authentication/authorization failed
- "internal_error" - Unexpected server error
- "timeout" - Operation timed out
- "database_error" - Database operation failed
```

### Example: What You'll See in Jaeger

**Success Case (Current - What you see):**
```
Logs (2)
├─ 71.2ms    event: user.id.generated
│            user.id: da6f6457-70d9-4584-a818-6d40cb25a121
│
└─ 211.85ms  event: user created successfully
             db.item_created: true
             user.id: da6f6457-70d9-4584-a818-6d40cb25a121
```

**Error Case (Before - What you see now):**
```
Logs (1)
└─ 150ms     exception: error occurred
             (stack trace, but not as clear)
```

**Error Case (After - What we want):**
```
Logs (2)
├─ 150ms     event: operation.failed
             error.type: database_error
             error.message: "dynamodb error put item"
             user.id: da6f6457-70d9-4584-a818-6d40cb25a121
             
└─ 150ms     exception: error occurred
             (stack trace for debugging)
```

**Benefits:**
- ✅ Clear error event in Logs section (like success events)
- ✅ Error type visible at a glance
- ✅ Error message visible
- ✅ Context (user.id, etc.) visible
- ✅ Easy to see what failed and why

## Implementation Plan

### Step 1: Enhance `RecordSpanError`
- Add error attributes
- Add error event
- Add error classification

### Step 2: Test in Jaeger
- Trigger various errors
- Verify visibility in Jaeger UI
- Check error filtering works

### Step 3: Optional Enhancements
- Add component/operation context (if needed)
- Add custom error attributes per error type

## Questions to Consider

1. Do we want to classify errors automatically, or manually specify type?
2. Should we add component/operation to every error, or is error type enough?
3. Do we want different error handling for different error types (4xx vs 5xx)?

## Expected Result

After implementation, when you look at a failed request in Jaeger:
- ✅ Error spans are clearly marked
- ✅ Can quickly see which span failed
- ✅ Error type is visible
- ✅ Can filter traces by error type
- ✅ Error details are easily accessible

