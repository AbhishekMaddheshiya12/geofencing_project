package api

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ctxKey int

const (
	ctxKeyStart ctxKey = iota
)

// Timing stashes the request start time in the context so handlers can
// compute the nanosecond elapsed value and embed it in the response.
func Timing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKeyStart, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// elapsedNS returns the nanoseconds since the request was received.
func elapsedNS(r *http.Request) int64 {
	if v := r.Context().Value(ctxKeyStart); v != nil {
		if t, ok := v.(time.Time); ok {
			return time.Since(t).Nanoseconds()
		}
	}
	return 0
}

// WriteJSON serializes payload as JSON, injecting `time_ns` as a string.
// payload must be a map[string]any so we can add the field without
// touching the caller's struct types.
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["time_ns"] = strconv.FormatInt(elapsedNS(r), 10)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("write json", "err", err)
	}
}

// WriteError is a convenience wrapper.
func WriteError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	WriteJSON(w, r, status, map[string]any{
		"error":   http.StatusText(status),
		"message": msg,
	})
}

// Logger logs every request with method, path, status, duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(ww, r)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker for WebSocket support
func (s *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := s.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Recover catches panics and turns them into 500 responses with time_ns.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic", "err", rec, "path", r.URL.Path)
				WriteError(w, r, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
