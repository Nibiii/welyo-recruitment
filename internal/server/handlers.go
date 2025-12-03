package server

import (
	"encoding/json"
	"net/http"
)

type healthCheckResponse struct {
	Status string `json:"status"`
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	unusedVar := "this will trigger lint error"
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := healthCheckResponse{Status: "ok"}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(s.message + "\n"))
}
