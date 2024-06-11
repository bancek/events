# events

HTTP event logger and middleware.

Before:

```go
http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello world"))
})

http.ListenAndServe("127.0.0.1:8080", nil)
```

After:

```go
import "github.com/bancek/events/v2"

http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello world"))
})

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

handler := events.NewHTTPMiddleware(http.DefaultServeMux, logger)

http.ListenAndServe("127.0.0.1:8080", handler)
```

Output:

```go
{"time":"2024-06-11T14:45:19.695142+02:00","level":"INFO","msg":"Event",
"httpSourceIp":"127.0.0.1","httpUserAgent":"curl/8.4.0","httpRespLen":11,
"requestId":"73152abd-a9fa-484f-bb75-659287254484","protocol":"http",
"httpMethod":"GET","httpUri":"/","httpRemoteAddr":"127.0.0.1:57309",
"duration":8000,"httpRespStatus":200}
```
