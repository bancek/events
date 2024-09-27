package events

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

var (
	AttrDuration       = "duration"
	AttrError          = "error"
	AttrErrorCause     = "errorCause"
	AttrErrorStack     = "errorStack"
	AttrHTTPMethod     = "httpMethod"
	AttrHTTPRemoteAddr = "httpRemoteAddr"
	AttrHTTPRespLen    = "httpRespLen"
	AttrHTTPRespStatus = "httpRespStatus"
	AttrHTTPSourceIP   = "httpSourceIp"
	AttrHTTPURI        = "httpUri"
	AttrHTTPUserAgent  = "httpUserAgent"
	AttrProtocol       = "protocol"
	AttrRequestID      = "requestId"
)

type Event struct {
	attrs map[string]slog.Value
	mutex sync.RWMutex

	baseLogger *slog.Logger
}

func NewEvent(logger *slog.Logger) *Event {
	requestID := uuid.New().String()

	e := &Event{
		attrs:      map[string]slog.Value{},
		baseLogger: logger,
	}

	e.SetAttr(AttrRequestID, requestID)

	return e
}

func (e *Event) SetAttr(key string, value any) {
	e.mutex.Lock()
	e.attrs[key] = slog.AnyValue(value)
	e.mutex.Unlock()
}

func (e *Event) GetAttr(key string) any {
	e.mutex.RLock()
	value := e.attrs[key].Any()
	e.mutex.RUnlock()

	return value
}

func (e *Event) SetError(err error) {
	e.SetAttr(AttrError, err)

	if _, ok := err.(xerrors.Wrapper); ok {
		e.SetAttr(AttrErrorStack, fmt.Sprintf("%+v", err))
	}

	if cause, ok := GetCause(err); ok {
		e.SetAttr(AttrErrorCause, fmt.Sprintf("%#v", cause))
	}
}

func (e *Event) Logger() *slog.Logger {
	e.mutex.RLock()

	attrs := make([]any, 0, len(e.attrs))

	for key, value := range e.attrs {
		attrs = append(attrs, slog.Attr{
			Key:   key,
			Value: value,
		})
	}

	e.mutex.RUnlock()

	return e.baseLogger.With(attrs...)
}
