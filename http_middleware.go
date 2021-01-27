package events

import (
	"net"
	"net/http"
	"strings"

	"github.com/koofr/go-httputils"
	"github.com/sirupsen/logrus"
)

type HTTPMiddleware struct {
	next   http.Handler
	logger *logrus.Entry

	BuildEvent         func(r *http.Request, e *Event)
	TrustForwardedFor  bool
	RequestIDHeaderKey string
	RequestLogMessage  string
	EventLogMessage    string
}

func NewHTTPMiddleware(next http.Handler, logger *logrus.Entry) *HTTPMiddleware {
	return &HTTPMiddleware{
		next:   next,
		logger: logger,

		BuildEvent:         nil,
		TrustForwardedFor:  true,
		RequestIDHeaderKey: "X-Request-Id",
		RequestLogMessage:  "HTTP request",
		EventLogMessage:    "Event",
	}
}

func (m *HTTPMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e := NewEvent(m.logger)

	e.SetField(FieldProtocol, "http")
	e.SetField(FieldHTTPMethod, r.Method)
	e.SetField(FieldHTTPURI, r.RequestURI)
	e.SetField(FieldHTTPRemoteAddr, r.RemoteAddr)
	e.SetField(FieldHTTPSourceIP, RequestIP(r, m.TrustForwardedFor))
	e.SetField(FieldHTTPUserAgent, r.UserAgent())

	if requestID := r.Header.Get(m.RequestIDHeaderKey); requestID != "" {
		e.SetField(FieldRequestID, requestID)
	}

	if m.BuildEvent != nil {
		m.BuildEvent(r, e)
	}

	e.Logger().Debug(m.RequestLogMessage)

	r = r.WithContext(NewEventContext(r.Context(), e))

	captureWriter := httputils.NewCaptureResponseWriter(w)

	m.next.ServeHTTP(captureWriter, r)

	e.SetField(FieldDuration, captureWriter.Duration())
	e.SetField(FieldHTTPRespStatus, captureWriter.StatusCode)
	e.SetField(FieldHTTPRespLen, captureWriter.ResponseLength)

	e.Logger().Info(m.EventLogMessage)
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
