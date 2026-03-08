package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	slog.Info("request received", "path", r.URL.Path, "method", r.Method)
	fmt.Fprintf(w, "Hello from testserver! Time: %s\n", time.Now().Format(time.RFC3339))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// Version changes demonstrate hot reload
func handleVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Version 1.0.0") // Change this to test hot reload
}
