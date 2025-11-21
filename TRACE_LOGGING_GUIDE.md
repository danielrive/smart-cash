# Trace Context in Logs - Implementation Guide

## Overview
Logs are now automatically correlated with traces using `trace_id` and `span_id` attributes. This allows you to:
- Find all logs for a specific trace
- Correlate logs with spans in your observability platform
- Debug issues by following a trace through all services

## Automatic Trace Context

### HTTP Request Logs
The HTTP middleware automatically adds trace context to all request logs:
```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "request processed",
  "method": "GET",
  "path": "/users/123",
  "status": 200,
  "duration_ms": "45ms",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

## Manual Trace Context (Services & Repositories)

To add trace context to logs in services and repositories, use the `WithTraceContext` helper:

### Before (No Trace Context)
```go
us.logger.Debug("getting user by id",
    slog.String("user_id", userId),
    slog.String("component", "service"),
)
```

### After (With Trace Context)
```go
us.logger.Debug("getting user by id",
    append([]any{
        slog.String("user_id", userId),
        slog.String("component", "service"),
    }, utils.WithTraceContext(ctx)...)...,
)
```

### Simpler Pattern (Using Helper)
You can also use the `LoggerWithTrace` helper:

```go
traceLogger := logging.LoggerWithTrace(ctx, us.logger)
traceLogger.Debug("getting user by id",
    slog.String("user_id", userId),
    slog.String("component", "service"),
)
```

## Example: Updated Service Method

```go
func (us *UserService) GetUserById(ctx context.Context, userId string) (models.UserResponse, error) {
    ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCGetUserById", "service",
        attribute.String("user.id", userId),
    )
    defer endSpan()

    // Option 1: Use LoggerWithTrace helper
    traceLogger := logging.LoggerWithTrace(ctx, us.logger)
    traceLogger.Debug("getting user by id",
        slog.String("user_id", userId),
        slog.String("component", "service"),
    )

    user, err := us.userRepository.GetUserById(ctx, userId)
    if err != nil {
        utils.RecordSpanError(ctx, err)
        traceLogger.Error("error getting user by id",
            slog.String("user_id", userId),
            slog.String("error", err.Error()),
            slog.String("component", "service"),
        )
        return models.UserResponse{}, err
    }

    traceLogger.Debug("user retrieved successfully",
        slog.String("user_id", userId),
        slog.String("component", "service"),
    )

    return user, nil
}
```

## Benefits

1. **Automatic Correlation**: HTTP request logs automatically include trace context
2. **Easy Integration**: Simple helper functions for manual trace context
3. **Observability**: All logs can be filtered by trace_id in your log aggregation system
4. **Debugging**: Follow a request through all services using trace_id

## Log Output Example

With trace context, your logs will look like:
```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "DEBUG",
  "msg": "getting user by id",
  "service": "user-service",
  "user_id": "123",
  "component": "service",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

## Next Steps

1. ✅ HTTP middleware logs include trace context automatically
2. 🔄 Update service/repository logs to include trace context (optional but recommended)
3. ✅ Configure your log aggregation system to index `trace_id` for easy filtering

