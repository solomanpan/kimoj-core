// Package rctx carries who is acting and the correlation id (per HTTP request / per consumed message).
package rctx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type RequestContext struct {
	Operator   string
	TraceID    string
	OrgUUID    string
	TenantUUID string
	UserUUID   string
}

type key struct{}

var anonymous = RequestContext{Operator: "system", TraceID: "-"}

func NewTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func With(ctx context.Context, rc RequestContext) context.Context {
	if rc.Operator == "" {
		rc.Operator = "system"
	}
	if rc.TraceID == "" {
		rc.TraceID = NewTraceID()
	}
	return context.WithValue(ctx, key{}, rc)
}

func From(ctx context.Context) RequestContext {
	if rc, ok := ctx.Value(key{}).(RequestContext); ok {
		return rc
	}
	return anonymous
}
