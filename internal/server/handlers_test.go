package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	server := &Server{
		logger: log.New(os.Stdout, "test: ", log.LstdFlags),
	}

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request returns 200 OK",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "POST request returns 405 Method Not Allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "PUT request returns 405 Method Not Allowed",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "DELETE request returns 405 Method Not Allowed",
			method:         "DELETE",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health-check", nil)
			w := httptest.NewRecorder()

			server.healthCheckHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.method == "GET" {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected content-type application/json, got %q", contentType)
				}
			}
		})
	}
}

func TestHelloWorldHandler(t *testing.T) {
	message := "Hello, World!"
	server := &Server{
		logger:  log.New(os.Stdout, "test: ", log.LstdFlags),
		message: message,
	}

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request returns 200 OK with custom message",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   message + "\n",
		},
		{
			name:           "POST request returns 405 Method Not Allowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "PUT request returns 405 Method Not Allowed",
			method:         "PUT",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/hello-world", nil)
			w := httptest.NewRecorder()

			server.helloWorldHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.method == "GET" {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("expected content-type text/plain, got %q", contentType)
				}
			}
		})
	}

	t.Run("different messages", func(t *testing.T) {
		customMessage := "Custom test message"
		customServer := &Server{
			logger:  log.New(os.Stdout, "test: ", log.LstdFlags),
			message: customMessage,
		}

		req := httptest.NewRequest("GET", "/hello-world", nil)
		w := httptest.NewRecorder()

		customServer.helloWorldHandler(w, req)

		body := w.Body.String()
		if body != customMessage+"\n" {
			t.Errorf("expected body %q, got %q", customMessage+"\n", body)
		}
	})
}

func TestLogMiddleware(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := log.New(&logBuffer, "test: ", log.LstdFlags)
	server := &Server{
		logger: logger,
	}

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("mock response"))
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}
	})

	middleware := server.logMiddleware(mockHandler)

	t.Run("middleware logs requests and calls next handler", func(t *testing.T) {
		logBuffer.Reset()

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body := w.Body.String()
		if body != "mock response" {
			t.Errorf("expected body %q, got %q", "mock response", body)
		}

		logOutput := logBuffer.String()
		if strings.Contains(logOutput, "GET /test") {
			t.Errorf("expected log to contain 'GET /test', got %q", logOutput)
		}
	})

	t.Run("middleware logs different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				logBuffer.Reset()

				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				logOutput := logBuffer.String()
				if !strings.Contains(logOutput, method+" /test") {
					t.Errorf("expected log to contain '%s /test', got %q", method, logOutput)
				}
			})
		}
	})

	t.Run("middleware preserves response headers", func(t *testing.T) {
		headerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "test-value")
			w.WriteHeader(http.StatusOK)
		})

		middlewareWithHeaders := server.logMiddleware(headerHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middlewareWithHeaders.ServeHTTP(w, req)

		resp := w.Result()
		headerValue := resp.Header.Get("X-Custom-Header")
		if headerValue != "test-value" {
			t.Errorf("expected header X-Custom-Header to be 'test-value', got %q", headerValue)
		}
	})
}

func BenchmarkHealthCheckHandler(b *testing.B) {
	server := &Server{
		logger: log.New(io.Discard, "", log.LstdFlags),
	}

	req := httptest.NewRequest("GET", "/health-check", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.healthCheckHandler(w, req)
	}
}

func BenchmarkHelloWorldHandler(b *testing.B) {
	server := &Server{
		logger:  log.New(io.Discard, "", log.LstdFlags),
		message: "Hello, World!",
	}

	req := httptest.NewRequest("GET", "/hello-world", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.helloWorldHandler(w, req)
	}
}

func BenchmarkLogMiddleware(b *testing.B) {
	server := &Server{
		logger: log.New(io.Discard, "", log.LstdFlags),
	}

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := server.logMiddleware(mockHandler)
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For header with single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Forwarded-For header with multiple IPs takes first",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1, 172.16.0.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Real-IP header when X-Forwarded-For is empty",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For takes precedence over X-Real-IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
				"X-Real-IP":       "203.0.113.1",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name:       "RemoteAddr when no headers present",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "10.0.0.1",
		},
		{
			name:       "RemoteAddr without port returns full address",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1",
			expectedIP: "10.0.0.1",
		},
		{
			name: "X-Forwarded-For with whitespace trimmed",
			headers: map[string]string{
				"X-Forwarded-For": "  192.168.1.100  ",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Real-IP with whitespace trimmed",
			headers: map[string]string{
				"X-Real-IP": "  203.0.113.1  ",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "203.0.113.1",
		},
		{
			name: "IPv6 address in X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1",
			},
			remoteAddr: "10.0.0.1:12345",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "IPv6 address in RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "[2001:db8::1]:12345",
			expectedIP: "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := getClientIP(req)

			if result != tt.expectedIP {
				t.Errorf("expected IP %q, got %q", tt.expectedIP, result)
			}
		})
	}
}
