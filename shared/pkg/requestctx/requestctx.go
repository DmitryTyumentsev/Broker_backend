package requestctx

import (
	"context"
	"strings"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ctx
	}

	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	requestID, ok := ctx.Value(requestIDContextKey).(string)
	requestID = strings.TrimSpace(requestID)

	return requestID, ok && requestID != ""
}
