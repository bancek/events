package events

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
)

type HTTPMiddleware struct {
	next   http.Handler
	logger *slog.Logger

	BuildEvent         func(r *http.Request, e *Event)
	TrustForwardedFor  bool
	RequestIDHeaderKey string
	RequestLogMessage  string
	EventLogMessage    string
	LogRequest         func(e *Event)
	LogEvent           func(e *Event)
}

func NewHTTPMiddleware(next http.Handler, logger *slog.Logger) *HTTPMiddleware {
	return &HTTPMiddleware{
		next:   next,
		logger: logger,

		BuildEvent:         nil,
		TrustForwardedFor:  true,
		RequestIDHeaderKey: "X-Request-Id",
		RequestLogMessage:  "HTTP request",
		EventLogMessage:    "Event",
		LogRequest:         nil,
		LogEvent:           nil,
	}
}

func (m *HTTPMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e := NewEvent(m.logger)

	ctx := NewEventContext(r.Context(), e)

	e.SetAttr(AttrProtocol, "http")
	e.SetAttr(AttrHTTPMethod, r.Method)
	e.SetAttr(AttrHTTPURI, r.RequestURI)
	e.SetAttr(AttrHTTPRemoteAddr, r.RemoteAddr)
	e.SetAttr(AttrHTTPSourceIP, RequestIP(r, m.TrustForwardedFor))
	e.SetAttr(AttrHTTPUserAgent, r.UserAgent())

	if requestID := r.Header.Get(m.RequestIDHeaderKey); requestID != "" {
		e.SetAttr(AttrRequestID, requestID)
	}

	if m.BuildEvent != nil {
		m.BuildEvent(r, e)
	}

	if m.LogRequest != nil {
		m.LogRequest(e)
	} else {
		e.Logger().LogAttrs(ctx, slog.LevelDebug, m.RequestLogMessage)
	}

	r = r.WithContext(ctx)

	captureWriter := NewCaptureResponseWriter(w)

	m.next.ServeHTTP(captureWriter, r)

	e.SetAttr(AttrDuration, captureWriter.Duration())
	e.SetAttr(AttrHTTPRespStatus, captureWriter.StatusCode)
	e.SetAttr(AttrHTTPRespLen, captureWriter.ResponseLength)

	if m.LogEvent != nil {
		m.LogEvent(e)
	} else {
		e.Logger().LogAttrs(r.Context(), slog.LevelInfo, m.EventLogMessage)
	}
}

func RequestIP(r *http.Request, trustForwardedFor bool) string {
	ip := r.Header.Get("X-Forwarded-For")

	if ip != "" && trustForwardedFor {
		if strings.ContainsRune(ip, ',') {
			ip = strings.Split(ip, ",")[0]
		}
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
}
