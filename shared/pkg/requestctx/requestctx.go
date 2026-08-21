package requestctx

import (
	"context"
	"net"
	"strings"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"
const clientIPContextKey contextKey = "client_ip"

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

func WithClientIP(ctx context.Context, clientIP string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return ctx
	}

	return context.WithValue(ctx, clientIPContextKey, ip.String())
}

func ClientIPFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	clientIP, ok := ctx.Value(clientIPContextKey).(string)
	clientIP = strings.TrimSpace(clientIP)

	return clientIP, ok && net.ParseIP(clientIP) != nil
}
