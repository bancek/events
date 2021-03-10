package events

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

var (
	FieldDuration       = "duration"
	FieldErrorCause     = "errorCause"
	FieldErrorStack     = "errorStack"
	FieldHTTPMethod     = "httpMethod"
	FieldHTTPRemoteAddr = "httpRemoteAddr"
	FieldHTTPRespLen    = "httpRespLen"
	FieldHTTPRespStatus = "httpRespStatus"
	FieldHTTPSourceIP   = "httpSourceIp"
	FieldHTTPURI        = "httpUri"
	FieldHTTPUserAgent  = "httpUserAgent"
	FieldProtocol       = "protocol"
	FieldRequestID      = "requestId"
)

type Event struct {
	fields logrus.Fields

	baseLogger *logrus.Entry
}

func NewEvent(logger *logrus.Entry) *Event {
	requestID := uuid.New().String()

	e := &Event{
		fields:     logrus.Fields{},
		baseLogger: logger,
	}

	e.SetField(FieldRequestID, requestID)

	return e
}

func (e *Event) SetField(key string, value interface{}) {
	e.fields[key] = value
}

func (e *Event) GetField(key string) interface{} {
	value, _ := e.fields[key]
	return value
}

func (e *Event) SetError(err error) {
	e.SetField(logrus.ErrorKey, err)

	if _, ok := err.(xerrors.Wrapper); ok {
		e.SetField(FieldErrorStack, fmt.Sprintf("%+v", err))
	}

	if cause, ok := GetCause(err); ok {
		e.SetField(FieldErrorCause, fmt.Sprintf("%#v", cause))
	}
}

func (e *Event) Logger() *logrus.Entry {
	return e.baseLogger.WithFields(e.fields)
}
