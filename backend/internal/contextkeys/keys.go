package contextkeys

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// TraceIDKey is the context key for trace ID
const TraceIDKey contextKey = "trace_id"
