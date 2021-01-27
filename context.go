package events

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

type contextKeyType string

var contextKey = contextKeyType("*events.Event")

func NewEventContext(ctx context.Context, event *Event) context.Context {
	return context.WithValue(ctx, contextKey, event)
}

func EventFromContext(ctx context.Context) *Event {
	return ctx.Value(contextKey).(*Event)
}

func SetRequestField(r *http.Request, key string, value interface{}) {
	e := EventFromContext(r.Context())
	e.SetField(key, value)
}

func SetRequestError(r *http.Request, err error) {
	e := EventFromContext(r.Context())
	e.SetError(err)
}

func ContextLogger(ctx context.Context) *logrus.Entry {
	e := EventFromContext(ctx)
	return e.Logger()
}

func RequestLogger(r *http.Request) *logrus.Entry {
	return ContextLogger(r.Context())
}
