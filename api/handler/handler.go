// Package handler implements the Wolllama API HTTP handlers.
package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	"github.com/wolllama/api/auth"
	"github.com/wolllama/api/db"
)

const (
	sessionName  = "wolllama-session"
	sessionKey   = "user_id"
	oauthStateKey = "oauth_state"
)

// Handler holds shared dependencies for HTTP handlers.
type Handler struct {
	db   *db.DB
	auth *auth.GitHubOAuth
	store *sessions.CookieStore
}

// New creates a Handler.
func New(database *db.DB, ghAuth *auth.GitHubOAuth) *Handler {
	// Session key from env or random (random means sessions reset on restart — fine for v1)
	key := []byte("wolllama-dev-key-change-in-production-32!")
	return &Handler{
		db:    database,
		auth:  ghAuth,
		store: sessions.NewCookieStore(key),
	}
}

// ---- Model handlers ----

// ListModels returns a paginated list of public models.
func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	offset := queryInt(r, "offset", 0)
	limit := queryInt(r, "limit", 20)
	search := r.URL.Query().Get("search")

	if limit > 100 {
		limit = 100
	}

	models, err := h.db.ListModels(offset, limit, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list models")
		slog.Error("list models", "error", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
		"offset": offset,
		"limit":  limit,
	})
}

// GetModel returns a single model by database ID.
func (h *Handler) GetModel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid model ID")
		return
	}

	model, err := h.db.GetModelByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get model")
		slog.Error("get model", "error", err)
		return
	}
	if model == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	writeJSON(w, http.StatusOK, model)
}

// SubmitModel accepts a new model submission.
func (h *Handler) SubmitModel(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	var req struct {
		ManifestObjID string  `json:"manifest_obj_id"`
		DisplayName   string  `json:"display_name"`
		DescriptionMd *string `json:"description_md"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.ManifestObjID == "" {
		writeError(w, http.StatusBadRequest, "manifest_obj_id is required")
		return
	}
	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "display_name is required")
		return
	}

	// TODO: fetch and validate Wolllama manifest from Walrus (sync validation)
	// For now, create the entry directly

	model := &db.Model{
		SubmitterID:   userID,
		ManifestObjID: req.ManifestObjID,
		DisplayName:   req.DisplayName,
		DescriptionMd: req.DescriptionMd,
	}

	if err := h.db.CreateModel(model); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			writeError(w, http.StatusConflict, "a model with this manifest object ID already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create model")
		slog.Error("create model", "error", err)
		return
	}

	writeJSON(w, http.StatusCreated, model)
}

// UserModels returns models published by a specific user.
func (h *Handler) UserModels(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	models, err := h.db.ListModelsByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list models")
		slog.Error("list user models", "error", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

// ---- Auth handlers ----

// Login redirects to GitHub OAuth.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	// Store state in a short-lived cookie for CSRF protection
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateKey,
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.auth.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback handles the GitHub OAuth callback.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// Verify state
	stateCookie, err := r.Cookie(oauthStateKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing state cookie")
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		writeError(w, http.StatusBadRequest, "invalid state")
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   oauthStateKey,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, err := h.auth.Exchange(ctx, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to exchange code")
		slog.Error("oauth exchange", "error", err)
		return
	}

	// Fetch GitHub user
	ghUser, err := h.auth.FetchUser(ctx, token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		slog.Error("fetch github user", "error", err)
		return
	}

	// Create or update user in DB
	user, err := h.db.CreateUser(ghUser.ID, ghUser.Login, &ghUser.AvatarURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		slog.Error("create user", "error", err)
		return
	}

	// Set session
	session, _ := h.store.Get(r, sessionName)
	session.Values[sessionKey] = user.ID
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		slog.Error("save session", "error", err)
		return
	}

	// Redirect to the site
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

// Me returns the currently authenticated user.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ---- Helpers ----

// requireAuth extracts the user ID from the session cookie. Returns 0 if not authenticated.
func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) int64 {
	session, _ := h.store.Get(r, sessionName)
	userID, ok := session.Values[sessionKey].(int64)
	if !ok || userID == 0 {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return 0
	}
	return userID
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}
