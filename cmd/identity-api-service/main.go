package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-app/internal/identity"
)

func main() {

	// 2. Init all dependencies
	container := identity.NewContainer()
	defer container.Close()

	// 2. Define Server Parameters using the container's router
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      container.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 3. Start Server Background Thread
	go func() {
		log.Printf("Identity Service API listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server stopped unexpectedly: %v", err)
		}
	}()

	// 4. Shutdown Trap
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	<-stopSignal // Block here until a signal is received
	log.Println("Shutdown signal intercepted. Clearing application tasks...")

	// Give active network requests 15 seconds to naturally finish processing
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to crash on shutdown: %v", err)
	}

	log.Println("Identity Service safely offline.")
}
