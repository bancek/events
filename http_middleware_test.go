package events_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"golang.org/x/xerrors"

	. "github.com/bancek/events"
)

var _ = Describe("HTTPMiddleware", func() {
	It("should log the event", func() {
		logger, hook := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		requestID := ""
		requestError := xerrors.Errorf("test error: %w", xerrors.Errorf("original error"))

		middleware := NewHTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID = EventFromContext(r.Context()).GetField(FieldRequestID).(string)
			time.Sleep(100 * time.Millisecond)
			SetRequestField(r, "customField", "custom value")
			SetRequestError(r, requestError)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("hello world"))
		}), logger.WithFields(nil))

		middleware.BuildEvent = func(r *http.Request, e *Event) {
			e.SetField("spanId", r.Header.Get("X-Span-Id"))
		}

		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "/path", nil)
		r.Header.Set("User-Agent", "TestUserAgent")
		r.Header.Set("X-Forwarded-For", "1.2.3.4")
		r.Header.Set("X-Span-Id", "123456789")

		middleware.ServeHTTP(w, r)

		Expect(w.Result().StatusCode).To(Equal(http.StatusCreated))
		Expect(w.Body.String()).To(Equal("hello world"))

		entries := hook.AllEntries()
		Expect(entries).To(HaveLen(2))
		Expect(entries[0].Level).To(Equal(logrus.DebugLevel))
		Expect(entries[0].Message).To(Equal("HTTP request"))
		Expect(entries[0].Data).To(Equal(logrus.Fields{
			"protocol":       "http",
			"httpMethod":     "POST",
			"httpUri":        "/path",
			"httpRemoteAddr": "192.0.2.1:1234",
			"httpSourceIp":   "1.2.3.4",
			"httpUserAgent":  "TestUserAgent",
			"requestId":      requestID,
			"spanId":         "123456789",
		}))
		Expect(entries[1].Level).To(Equal(logrus.InfoLevel))
		Expect(entries[1].Message).To(Equal("Event"))
		eventFields := entries[1].Data
		Expect(eventFields).To(Equal(logrus.Fields{
			"protocol":       "http",
			"httpMethod":     "POST",
			"httpUri":        "/path",
			"httpRemoteAddr": "192.0.2.1:1234",
			"httpSourceIp":   "1.2.3.4",
			"httpUserAgent":  "TestUserAgent",
			"requestId":      requestID,
			"spanId":         "123456789",
			"duration":       eventFields["duration"],
			"httpRespStatus": 201,
			"httpRespLen":    int64(11),
			"error":          requestError,
			"errorCause":     eventFields["errorCause"],
			"errorStack":     eventFields["errorStack"],
			"customField":    "custom value",
		}))
		Expect(eventFields["errorCause"]).To(ContainSubstring("original error"))
		Expect(eventFields["errorStack"]).To(ContainSubstring("http_middleware_test.go"))
		Expect(eventFields["duration"]).To(BeNumerically(">=", 100*time.Millisecond))
	})
})
