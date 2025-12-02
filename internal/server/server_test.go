package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	t.Run("creates server with provided logger and message", func(t *testing.T) {
		message := "Test Message"
		var logBuffer bytes.Buffer
		logger := log.New(&logBuffer, "custom: ", log.LstdFlags)

		server := NewServer(message, logger)

		if server == nil {
			t.Fatal("expected server to be created, got nil")
		}

		if server.message != message {
			t.Errorf("expected message %q, got %q", message, server.message)
		}

		if server.logger != logger {
			t.Error("expected provided logger to be used")
		}

		if server.mux == nil {
			t.Error("expected mux to be initialized")
		}

		handler := server.Handler()
		if handler == nil {
			t.Error("expected handler to be returned")
		}

		if handler != server.mux {
			t.Error("expected handler to be the mux")
		}
	})

	t.Run("creates server with default logger when nil is provided", func(t *testing.T) {
		message := "Test Message"

		server := NewServer(message, nil)

		if server == nil {
			t.Fatal("expected server to be created, got nil")
		}

		if server.logger == nil {
			t.Error("expected default logger to be set")
		}

		if server.message != message {
			t.Errorf("expected message %q, got %q", message, server.message)
		}
	})

	t.Run("server has routes configured after creation", func(t *testing.T) {
		message := "Test Message"
		logger := log.New(os.Stdout, "test: ", log.LstdFlags)

		server := NewServer(message, logger)

		req := httptest.NewRequest("GET", "/health-check", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected health-check route to be configured, got status %d", resp.StatusCode)
		}

		req2 := httptest.NewRequest("GET", "/hello-world", nil)
		w2 := httptest.NewRecorder()

		server.Handler().ServeHTTP(w2, req2)

		resp2 := w2.Result()
		if resp2.StatusCode != http.StatusOK {
			t.Errorf("expected hello-world route to be configured, got status %d", resp2.StatusCode)
		}

		body := w2.Body.String()
		if body != message+"\n" {
			t.Errorf("expected message %q to be used in hello-world handler, got %q", message+"\n", body)
		}
	})
}

func TestServerRoutes(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := log.New(&logBuffer, "test: ", log.LstdFlags)
	message := "Test Hello Message"

	server := NewServer(message, logger)
	handler := server.Handler()

	t.Run("health-check route", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("GET", "/health-check", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var healthResp healthCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &healthResp)
		if err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}

		if healthResp.Status != "ok" {
			t.Errorf("expected status 'ok', got %q", healthResp.Status)
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "GET /health-check") {
			t.Errorf("expected log to contain 'GET /health-check', got %q", logOutput)
		}
	})

	t.Run("hello-world route", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("GET", "/hello-world", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body := w.Body.String()
		if body != message+"\n" {
			t.Errorf("expected body %q, got %q", message+"\n", body)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("expected content-type text/plain, got %q", contentType)
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "GET /hello-world") {
			t.Errorf("expected log to contain 'GET /hello-world', got %q", logOutput)
		}
	})

	t.Run("invalid route returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/invalid-route", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("health-check with POST method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/health-check", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "POST /health-check") {
			t.Errorf("expected log to contain 'POST /health-check', got %q", logOutput)
		}
	})

	t.Run("hello-world with POST method", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("POST", "/hello-world", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "POST /hello-world") {
			t.Errorf("expected log to contain 'POST /hello-world', got %q", logOutput)
		}
	})
}

func BenchmarkServerRoutes(b *testing.B) {
	server := NewServer("Hello, World!", log.New(io.Discard, "", log.LstdFlags))
	handler := server.Handler()

	b.Run("health-check", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/health-check", nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})

	b.Run("hello-world", func(b *testing.B) {
		req := httptest.NewRequest("GET", "/hello-world", nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}
