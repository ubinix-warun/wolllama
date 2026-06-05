package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/wolllama/api/auth"
	"github.com/wolllama/api/db"
	"github.com/wolllama/api/handler"
)

// siteFS embeds the built React SPA for production deployments.
// In development, this embed will be empty (site/dist only gets populated
// when the build script copies ../site/dist into api/site/dist).
//
//go:embed site/dist/*
var siteFS embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Database
	database, err := db.Open("wolllama.db")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Auth
	ghAuth, err := auth.NewGitHubOAuth()
	if err != nil {
		slog.Error("failed to configure GitHub OAuth", "error", err)
		os.Exit(1)
	}

	// Handlers
	h := handler.New(database, ghAuth)

	// Routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/models", h.ListModels)
	mux.HandleFunc("GET /api/models/{id}", h.GetModel)
	mux.HandleFunc("POST /api/models", h.SubmitModel)
	mux.HandleFunc("GET /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/auth/callback", h.Callback)
	mux.HandleFunc("GET /api/auth/me", h.Me)
	mux.HandleFunc("GET /api/users/{id}/models", h.UserModels)

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Serve embedded SPA if available, otherwise fall back to dev message
	if hasEmbeddedSite() {
		slog.Info("serving embedded SPA from siteFS")
		spaFS, err := fs.Sub(siteFS, "site/dist")
		if err != nil {
			slog.Error("failed to sub siteFS", "error", err)
		} else {
			fileServer := http.FileServer(http.FS(spaFS))
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				// SPA fallback: serve index.html for unknown paths
				if _, err := spaFS.Open(r.URL.Path); err != nil {
					r.URL.Path = "/"
				}
				fileServer.ServeHTTP(w, r)
			})
		}
	} else {
		slog.Info("no embedded site — running API-only mode")
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><title>Wolllama API</title></head>
<body style="font-family:monospace;background:#0d0d0d;color:#e0e0e0;padding:40px">
<h1>Wolllama API</h1>
<p>The API is running. In development, start the site separately:</p>
<pre style="background:#1a1a2e;padding:16px;border-radius:8px">
cd site && npm run dev
</pre>
</body></html>`)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("starting wolllama API", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// hasEmbeddedSite checks whether the siteFS embed contains actual files.
func hasEmbeddedSite() bool {
	entries, err := siteFS.ReadDir("site/dist")
	if err != nil {
		return false
	}
	return len(entries) > 0
}
