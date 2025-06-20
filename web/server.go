package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// Serve static files from web/ (e.g., index.html)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))

	// Serve game binary from parent directory
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		binaryPath := "../bin/main"
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			http.Error(w, "Game binary not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=main")
		http.ServeFile(w, r, binaryPath)
	})

	// Serve index.html at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	})

	// Start server on port 8080 with timeout settings
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Printf("Server starting on http://localhost:8080\n")
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
