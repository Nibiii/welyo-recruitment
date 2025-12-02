package server

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type Server struct {
	mux     *http.ServeMux
	logger  *log.Logger
	message string
}

func NewServer(message string, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.Default()
	}

	s := &Server{
		mux:     http.NewServeMux(),
		logger:  logger,
		message: message,
	}

	s.routes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.Handle("/health-check", s.logMiddleware(http.HandlerFunc(s.healthCheckHandler)))
	s.mux.Handle("/hello-world", s.logMiddleware(http.HandlerFunc(s.helloWorldHandler)))
}

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		s.logger.Printf("%s %s - %d - %s - %v",
			r.Method, r.URL.Path, wrapped.statusCode, getClientIP(r), duration)
	})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (common nginx/cloudflare setup)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
