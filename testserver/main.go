package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
)

var version = "1.0.0"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Hot Reload Demo Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; color: #333; }
        .info { background: #e8f4fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .stat { background: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; }
        .footer { text-align: center; margin-top: 30px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="header">🔥 Hot Reload Demo Server</h1>
        
        <div class="info">
            <strong>This server is being hot-reloaded!</strong><br>
            Edit any .go file and watch the server restart automatically.
        </div>
        
        <div class="stats">
            <div class="stat">
                <h3>Version</h3>
                <p>%s</p>
            </div>
            <div class="stat">
                <h3>Current Time</h3>
                <p>%s</p>
            </div>
            <div class="stat">
                <h3>Request Method</h3>
                <p>%s</p>
            </div>
            <div class="stat">
                <h3>User Agent</h3>
                <p>%s</p>
            </div>
        </div>
        
        <div class="info">
            <h3>How it works:</h3>
            <ol>
                <li>File watcher detects .go file changes</li>
                <li>Debouncer prevents multiple rapid builds</li>
                <li>Builder compiles the new version</li>
                <li>Process manager stops old server</li>
                <li>New server starts with updated code</li>
            </ol>
        </div>
        
        <div class="footer">
            <p>Built with Go | Hot Reload Engine | Version %s</p>
        </div>
    </div>
</body>
</html>`, version, time.Now().Format("2006-01-02 15:04:05"), r.Method, r.UserAgent(), version)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","version":"%s","time":"%s"}`, version, time.Now().Format(time.RFC3339))
	})

	http.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"server": "Hot Reload Demo",
			"version": "%s",
			"go_version": "%s",
			"uptime": "%s",
			"request_count": %d
		}`, version, runtime.Version(), time.Since(time.Now()).String(), 42)
	})

	log.Printf("🚀 Demo server starting on http://localhost:%s", port)
	log.Printf("📊 Health check: http://localhost:%s/health", port)
	log.Printf("🔥 Hot reload enabled - edit .go files to restart!")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
