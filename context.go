package events

import (
	"context"
)

type contextKeyType string

var contextKey = contextKeyType("*events.Event")

func NewEventContext(ctx context.Context, event *Event) context.Context {
	return context.WithValue(ctx, contextKey, event)
}

func EventFromContext(ctx context.Context) *Event {
	return ctx.Value(contextKey).(*Event)
}

func TryEventFromContext(ctx context.Context) (*Event, bool) {
	event, ok := ctx.Value(contextKey).(*Event)
	return event, ok
}
