package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/wolllama/api/auth"
	"github.com/wolllama/api/db"
	"github.com/wolllama/api/handler"
	wwalrus "github.com/wolllama/pkg/walrus"
)

// siteFS embeds the built React SPA for production deployments.
// In development, this embed will be empty (site/dist only gets populated
// when the build script copies ../site/dist into api/site/dist).
//
//go:embed site/dist
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

	// Auth mode: "open" (no auth), "token" (bearer token), or "github" (OAuth)
	authMode := handler.AuthMode(os.Getenv("WOLLLAMA_AUTH_MODE"))
	switch authMode {
	case handler.AuthModeOpen:
		slog.Info("auth mode: open (no authentication required)")
	case handler.AuthModeSui:
		slog.Info("auth mode: sui (wallet signature on submission)")
	case handler.AuthModeToken:
		slog.Info("auth mode: token (bearer token)")
	default:
		authMode = handler.AuthModeOpen
		slog.Info("auth mode: open (default)")
	}

	// GitHub OAuth (only initialized in github mode)
	var ghAuth *auth.GitHubOAuth
	if authMode == handler.AuthModeGitHub {
		ghAuth, _ = auth.NewGitHubOAuth()
		if ghAuth == nil {
			slog.Warn("GitHub OAuth not configured — falling back to open mode")
			authMode = handler.AuthModeOpen
		}
	}

	apiToken := os.Getenv("WOLLLAMA_API_TOKEN")
	if authMode == handler.AuthModeToken && apiToken == "" {
		slog.Warn("token mode enabled but WOLLLAMA_API_TOKEN not set — generating a random token")
		apiToken = randomToken()
		slog.Info("use this token for API requests", "token", apiToken)
	}

	// Walrus client — respects WOLLLAMA_WALRUS_NETWORK (mainnet/testnet)
	network := os.Getenv("WOLLLAMA_WALRUS_NETWORK")
	aggURL := os.Getenv("WOLLLAMA_AGGREGATOR_URL")

	switch {
	case aggURL != "":
		// explicit URL overrides everything
	case network == "mainnet":
		aggURL = "https://aggregator.walrus-mainnet.walrus.space"
	default:
		aggURL = "https://aggregator.walrus-testnet.walrus.space"
	}
	slog.Info("walrus aggregator", "url", aggURL)

	walrusClient := wwalrus.NewClient(wwalrus.Config{
		AggregatorURLs: []string{aggURL},
	})

	// Handlers
	featuredOwners := strings.Split(os.Getenv("WOLLLAMA_FEATURED_OWNERS"), ",")

	h := handler.New(database, ghAuth, authMode, apiToken, walrusClient, featuredOwners)

	// Routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/blobs/{obj_id}", h.GetBlobContent)
	mux.HandleFunc("GET /api/manifest/preview", h.PreviewManifest)
	mux.HandleFunc("GET /api/models/featured", h.ListFeaturedModels)
	mux.HandleFunc("PUT /api/models/{id}/featured", h.ToggleFeatured)
	mux.HandleFunc("GET /api/models", h.ListModels)
	mux.HandleFunc("GET /api/models/{id}", h.GetModel)
	mux.HandleFunc("POST /api/models", h.SubmitModel)
	mux.HandleFunc("GET /api/auth/sui/nonce", h.SuiNonce)
	mux.HandleFunc("POST /api/auth/sui/verify", h.SuiVerify)
	mux.HandleFunc("GET /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/auth/callback", h.Callback)
	mux.HandleFunc("GET /api/auth/me", h.Me)
	mux.HandleFunc("GET /api/users/{id}/models", h.UserModels)

	// Health check
	// Config endpoint — tells the frontend which network to use
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		suiRPC := os.Getenv("WOLLLAMA_SUI_RPC_URL")
		addr := r.URL.Query().Get("address")
		isOwner := "false"
		if addr != "" {
			for _, o := range strings.Split(os.Getenv("WOLLLAMA_FEATURED_OWNERS"), ",") {
				if strings.TrimSpace(o) == addr {
					isOwner = "true"
					break
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"walrus_network":"%s","sui_network":"%s","sui_rpc_url":"%s","is_featured_owner":%s}`,
			os.Getenv("WOLLLAMA_WALRUS_NETWORK"),
			os.Getenv("WOLLLAMA_SUI_NETWORK"),
			suiRPC,
			isOwner,
		)
	})

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
				// SPA fallback: serve index.html for paths without a matching file.
				// fs.FS paths must not start with "/" — strip before checking.
				path := strings.TrimPrefix(r.URL.Path, "/")
				if path == "" {
					path = "."
				}
				if _, err := spaFS.Open(path); err != nil {
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

func randomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// hasEmbeddedSite checks whether the siteFS embed contains actual files.
func hasEmbeddedSite() bool {
	entries, err := siteFS.ReadDir("site/dist")
	if err != nil {
		return false
	}
	return len(entries) > 0
}
