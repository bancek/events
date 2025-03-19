package events_test

import (
	"context"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"

	. "github.com/bancek/events/v2"
)

var _ = Describe("HTTPMiddleware", func() {
	It("should log the event", func() {
		testRecords := &testRecords{}
		logger := slog.New(&testHandler{testRecords: testRecords, attrsMap: map[string]any{}})

		requestID := ""
		requestError := xerrors.Errorf("test error: %w", xerrors.Errorf("original error"))

		middleware := NewHTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			event := EventFromContext(r.Context())
			requestID = event.GetAttr(AttrRequestID).(string)
			time.Sleep(100 * time.Millisecond)
			event.SetAttr("customAttr", "custom value")
			event.SetError(requestError)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("hello world"))
		}), logger)

		middleware.BuildEvent = func(r *http.Request, e *Event) {
			e.SetAttr("spanId", r.Header.Get("X-Span-Id"))
		}

		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "/path", nil)
		r.Header.Set("User-Agent", "TestUserAgent")
		r.Header.Set("X-Forwarded-For", "1.2.3.4")
		r.Header.Set("X-Span-Id", "123456789")

		middleware.ServeHTTP(w, r)

		Expect(w.Result().StatusCode).To(Equal(http.StatusCreated))
		Expect(w.Body.String()).To(Equal("hello world"))

		Expect(testRecords.Records).To(HaveLen(2))
		Expect(testRecords.Records[0].Level).To(Equal(slog.LevelDebug))
		Expect(testRecords.Records[0].Message).To(Equal("HTTP request"))
		Expect(testRecords.Records[0].AttrsMap).To(Equal(map[string]any{
			"protocol":       "http",
			"httpMethod":     "POST",
			"httpUri":        "/path",
			"httpRemoteAddr": "192.0.2.1:1234",
			"httpSourceIp":   "1.2.3.4",
			"httpUserAgent":  "TestUserAgent",
			"requestId":      requestID,
			"spanId":         "123456789",
		}))
		Expect(testRecords.Records[1].Level).To(Equal(slog.LevelInfo))
		Expect(testRecords.Records[1].Message).To(Equal("Event"))
		eventAttrs := testRecords.Records[1].AttrsMap
		Expect(eventAttrs).To(Equal(map[string]any{
			"protocol":       "http",
			"httpMethod":     "POST",
			"httpUri":        "/path",
			"httpRemoteAddr": "192.0.2.1:1234",
			"httpSourceIp":   "1.2.3.4",
			"httpUserAgent":  "TestUserAgent",
			"requestId":      requestID,
			"spanId":         "123456789",
			"duration":       eventAttrs["duration"],
			"httpRespStatus": int64(201),
			"httpRespLen":    int64(11),
			"error":          requestError,
			"errorCause":     eventAttrs["errorCause"],
			"errorStack":     eventAttrs["errorStack"],
			"customAttr":     "custom value",
		}))
		Expect(eventAttrs["errorCause"]).To(ContainSubstring("original error"))
		Expect(eventAttrs["errorStack"]).To(ContainSubstring("http_middleware_test.go"))
		Expect(eventAttrs["duration"]).To(BeNumerically(">=", 100*time.Millisecond))
	})

	It("should log the event with custom hooks", func() {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		requestLogged := false
		eventLogged := false

		middleware := NewHTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(requestLogged).To(BeTrue())
			Expect(eventLogged).To(BeFalse())
		}), logger)

		middleware.LogRequest = func(ctx context.Context, e *Event) {
			requestLogged = true
		}

		middleware.LogEvent = func(ctx context.Context, e *Event) {
			eventLogged = true
		}

		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, httptest.NewRequest("POST", "/path", nil))

		Expect(eventLogged).To(BeTrue())
	})
})

type testRecord struct {
	slog.Record
	AttrsMap map[string]any
}

type testRecords struct {
	Records []*testRecord
}

type testHandler struct {
	testRecords *testRecords
	attrsMap    map[string]any
}

func (h *testHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := maps.Clone(h.attrsMap)
	for _, attr := range attrs {
		newAttrs[attr.Key] = attr.Value.Any()
	}
	return &testHandler{
		testRecords: h.testRecords,
		attrsMap:    newAttrs,
	}
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return &testHandler{
		testRecords: h.testRecords,
		attrsMap:    maps.Clone(h.attrsMap),
	}
}

func (h *testHandler) Handle(ctx context.Context, record slog.Record) error {
	r := &testRecord{
		Record:   record,
		AttrsMap: maps.Clone(h.attrsMap),
	}
	record.Attrs(func(a slog.Attr) bool {
		r.AttrsMap[a.Key] = a.Value.Any()
		return true
	})
	h.testRecords.Records = append(h.testRecords.Records, r)
	return nil
}
