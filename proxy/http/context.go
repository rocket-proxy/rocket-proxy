package http

import (
	"context"
	"net/http"
)

const (
	ctxKeyHttpResponseWriter = "http.response-writer"
	ctxKeyHttpHttpRequest    = "http.response-request"
)

func setWithUserContext(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	ctx = context.WithValue(ctx, ctxKeyHttpResponseWriter, w)
	ctx = context.WithValue(ctx, ctxKeyHttpHttpRequest, r)
	return ctx
}

func requiredResponseWriter(ctx context.Context) http.ResponseWriter {
	v, ok := ctx.Value(ctxKeyHttpResponseWriter).(http.ResponseWriter)
	if ok {
		return v
	}
	panic("ResponseWriter not in context")
}

func requiredHttpRequest(ctx context.Context) *http.Request {
	v, ok := ctx.Value(ctxKeyHttpHttpRequest).(*http.Request)
	if ok {
		return v
	}
	panic("*http.Request not in context")
}
