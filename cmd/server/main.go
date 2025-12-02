package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"welyo/recruitment/internal/server"
)

func main() {
	hello := os.Getenv("SERVER_HELLO")
	if hello == "" {
		log.Fatal("SERVER_HELLO env variable must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	srv := http.Server{
		Addr:         ":" + port,
		Handler:      server.NewServer(hello, logger).Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Println("Starting server on port", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not start server: %v", err)
	}
}
