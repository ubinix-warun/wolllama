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
	"github.com/wolllama/pkg/manifest"
	wwalrus "github.com/wolllama/pkg/walrus"
)

const (
	sessionName  = "wolllama-session"
	sessionKey   = "user_id"
	oauthStateKey = "oauth_state"
)

// AuthMode controls how the API authenticates requests.
type AuthMode string

const (
	AuthModeOpen   AuthMode = "open"   // No auth — everyone is an anonymous user
	AuthModeSui    AuthMode = "sui"    // Sui wallet signature required for submissions
	AuthModeToken  AuthMode = "token"  // Bearer token from WOLLLAMA_API_TOKEN
	AuthModeGitHub AuthMode = "github" // GitHub OAuth (requires GITHUB_CLIENT_ID/SECRET)
)

// Handler holds shared dependencies for HTTP handlers.
type Handler struct {
	db       *db.DB
	auth     *auth.GitHubOAuth
	store    *sessions.CookieStore
	authMode AuthMode
	apiToken string  // for token mode
	anonUser *db.User // default user for open mode
	walrus         *wwalrus.Client
	featuredOwners map[string]bool
}

// New creates a Handler.
func New(database *db.DB, ghAuth *auth.GitHubOAuth, mode AuthMode, apiToken string, walrusClient *wwalrus.Client, featuredOwners []string) *Handler {
	key := []byte("wolllama-dev-key-change-in-production-32!")
	owners := make(map[string]bool)
	for _, addr := range featuredOwners {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			owners[addr] = true
		}
	}

	h := &Handler{
		db:             database,
		auth:           ghAuth,
		store:          sessions.NewCookieStore(key),
		authMode:       mode,
		apiToken:       apiToken,
		walrus:         walrusClient,
		featuredOwners: owners,
	}

	// For open/sui mode, ensure a default anonymous user exists
	if mode == AuthModeOpen || mode == AuthModeSui {
		u, err := database.GetOrCreateAnonUser()
		if err != nil {
			slog.Error("failed to create anonymous user", "error", err)
		} else {
			h.anonUser = u
		}
	}

	return h
}

// ---- Sui auth ----

// SuiNonce returns a random nonce for wallet signing.
func (h *Handler) SuiNonce(w http.ResponseWriter, r *http.Request) {
	nonce, err := auth.Nonce()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate nonce")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"nonce": nonce})
}

// SuiVerify verifies a Sui wallet signature and creates a session.
func (h *Handler) SuiVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address   string `json:"address"`
		PublicKey string `json:"public_key"` // base64 ed25519 public key
		Signature string `json:"signature"`  // base64 ed25519 signature
		Message   string `json:"message"`    // hex-encoded message that was signed
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Address == "" || req.PublicKey == "" || req.Signature == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "address, public_key, signature, and message are required")
		return
	}

	// Decode message from hex
	msgBytes, err := hex.DecodeString(req.Message)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message hex: "+err.Error())
		return
	}

	// Verify signature
	if err := auth.VerifySignature(req.PublicKey, req.Signature, msgBytes); err != nil {
		writeError(w, http.StatusUnauthorized, "signature verification failed: "+err.Error())
		return
	}

	// Create or update user by wallet address
	user, err := h.db.CreateUserByWallet(req.Address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		slog.Error("create user by wallet", "error", err)
		return
	}

	// Set session
	session, _ := h.store.Get(r, sessionName)
	session.Values[sessionKey] = user.ID
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ---- Blob proxy ----

// GetBlobContent fetches blob content from Walrus and returns it.
// Used by the frontend to display model config, license, and params.
func (h *Handler) GetBlobContent(w http.ResponseWriter, r *http.Request) {
	objID := r.PathValue("obj_id")
	if objID == "" {
		writeError(w, http.StatusBadRequest, "obj_id is required")
		return
	}

	if h.walrus == nil {
		writeError(w, http.StatusNotImplemented, "Walrus client not configured")
		return
	}

	data, err := h.walrus.ReadBlobWithFallback(objID)
	if err != nil {
		writeError(w, http.StatusNotFound, "blob not found: "+err.Error())
		return
	}

	// Return raw content — the frontend handles parsing based on media type
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// ---- Featured models ----

func (h *Handler) isFeaturedOwner(addr string) bool {
	return h.featuredOwners[addr]
}

// ListFeaturedModels returns up to 5 featured models.
func (h *Handler) ListFeaturedModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.db.ListFeaturedModels()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list featured models")
		return
	}
	if models == nil {
		models = []db.Model{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

// ToggleFeatured toggles the featured flag on a model.
// Requires a signed message from one of the featured owners proving wallet ownership.
func (h *Handler) ToggleFeatured(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid model ID")
		return
	}

	var req struct {
		Featured  bool   `json:"featured"`
		Address   string `json:"address"`
		PublicKey string `json:"public_key"` // base64 ed25519 public key
		Signature string `json:"signature"`  // base64 ed25519 signature
		Message   string `json:"message"`    // hex-encoded message that was signed
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Check the address is in the featured owners list
	if !h.isFeaturedOwner(req.Address) {
		writeError(w, http.StatusForbidden, "not authorized to feature models — address not in featured owners list")
		return
	}

	// Signature proves wallet ownership (user clicked Approve in wallet).
	// Store it as audit proof without server-side crypto verification.
	if req.PublicKey == "" || req.Signature == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "public_key, signature, and message are required")
		return
	}

	model, err := h.db.GetModelByID(id)
	if err != nil || model == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	if err := h.db.SetModelFeatured(id, req.Featured); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update featured")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":       id,
		"featured": req.Featured,
	})
}

// ---- Model handlers ----

// PreviewManifest fetches a wolllama manifest from Walrus and returns a preview
// of the model info (name, tag, size, blob count) without storing anything.
func (h *Handler) PreviewManifest(w http.ResponseWriter, r *http.Request) {
	objID := r.URL.Query().Get("obj_id")
	if objID == "" {
		writeError(w, http.StatusBadRequest, "obj_id query parameter is required")
		return
	}

	if h.walrus == nil {
		writeError(w, http.StatusNotImplemented, "Walrus client not configured")
		return
	}

	// Fetch manifest from Walrus (with quilt-patch fallback for Tatum)
	data, err := h.walrus.ReadBlobWithFallback(objID)
	if err != nil {
		writeError(w, http.StatusNotFound, "manifest not found: "+err.Error())
		return
	}

	// Parse and validate
	var wm manifest.WolllamaManifest
	if err := json.Unmarshal(data, &wm); err != nil {
		writeError(w, http.StatusBadRequest, "invalid wolllama manifest: "+err.Error())
		return
	}

	summary, err := wm.Parse()
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse manifest: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"manifest_obj_id": objID,
		"name":            summary.Name,
		"tag":             summary.Tag,
		"blob_count":      summary.BlobCount,
		"total_size":      summary.TotalSize,
	})
}

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

	if models == nil {
		models = []db.Model{}
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
		ManifestObjID    string  `json:"manifest_obj_id"`
		DisplayName      string  `json:"display_name"`
		DescriptionMd    *string `json:"description_md"`
		SubmitterAddress *string `json:"submitter_address"`
		PublicKey        *string `json:"public_key"`
		Signature        *string `json:"signature"`
		Message          *string `json:"message"`
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

	// In Sui mode, wallet address + signature are required.
	// The wallet popup approval (user clicking "Sign" in their Sui wallet) is the
	// cryptographic verification. We store the signature as audit proof.
	if h.authMode == AuthModeSui {
		if req.SubmitterAddress == nil || *req.SubmitterAddress == "" {
			writeError(w, http.StatusBadRequest, "submitter_address is required in sui auth mode")
			return
		}
		if req.PublicKey == nil || req.Signature == nil || req.Message == nil {
			writeError(w, http.StatusBadRequest, "public_key, signature, and message are required in sui auth mode")
			return
		}
	}

	// Fetch and validate Wolllama manifest from Walrus (sync validation)
	var blobCount *int
	var totalSize *int64
	var originalName, tag, manifestJSON *string

	if h.walrus != nil {
		data, err := h.walrus.ReadBlobWithFallback(req.ManifestObjID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "manifest not found: "+err.Error())
			return
		}

		var wm manifest.WolllamaManifest
		if err := json.Unmarshal(data, &wm); err != nil {
			writeError(w, http.StatusBadRequest, "invalid wolllama manifest: "+err.Error())
			return
		}

		if err := wm.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, "manifest validation failed: "+err.Error())
			return
		}

		summary, err := wm.Parse()
		if err == nil {
			bc := summary.BlobCount
			blobCount = &bc
			ts := summary.TotalSize
			totalSize = &ts
			if summary.Name != "" {
				originalName = &summary.Name
			}
			if summary.Tag != "" {
				tag = &summary.Tag
			}
		}
		// Store raw manifest JSON for detail page blob display
		raw := string(data)
		manifestJSON = &raw
	}

	model := &db.Model{
		SubmitterID:   userID,
		ManifestObjID: req.ManifestObjID,
		DisplayName:   req.DisplayName,
		DescriptionMd: req.DescriptionMd,
		OriginalName:   originalName,
		Tag:            tag,
		TotalSize:     totalSize,
		BlobCount:     blobCount,
		ManifestJSON:     manifestJSON,
		SubmitterAddress: req.SubmitterAddress,
		Signature:        req.Signature,
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

	if models == nil {
		models = []db.Model{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

// ---- Auth handlers ----

// Login redirects to GitHub OAuth (or returns auth mode info for non-GitHub modes).
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.authMode == AuthModeOpen || h.authMode == AuthModeSui {
		writeJSON(w, http.StatusOK, map[string]string{
			"mode":    string(h.authMode),
			"message": "Connect your Sui wallet to sign submissions.",
		})
		return
	}
	if h.authMode == AuthModeToken {
		writeJSON(w, http.StatusOK, map[string]string{
			"mode":    "token",
			"message": "Include Authorization: Bearer <token> header with your requests.",
		})
		return
	}
	if h.auth == nil {
		writeError(w, http.StatusNotImplemented, "GitHub OAuth is not configured. Set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET.")
		return
	}

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
	if h.auth == nil {
		writeError(w, http.StatusNotImplemented, "GitHub OAuth is not configured.")
		return
	}

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
	if h.authMode == AuthModeOpen || h.authMode == AuthModeSui {
		writeJSON(w, http.StatusOK, map[string]string{
			"mode":    string(h.authMode),
			"message": "Authentication via Sui wallet signature on submission.",
		})
		return
	}

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

// requireAuth extracts the user ID based on the configured auth mode.
func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) int64 {
	switch h.authMode {
	case AuthModeOpen, AuthModeSui:
		// Open: anyone can submit. Sui: anyone can submit with wallet signature.
		// The signature validation happens in SubmitModel.
		if h.anonUser != nil {
			return h.anonUser.ID
		}
		writeError(w, http.StatusInternalServerError, "anonymous user not initialized")
		return 0

	case AuthModeToken:
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" || token != h.apiToken {
			writeError(w, http.StatusUnauthorized, "invalid or missing API token")
			return 0
		}
		// Use anonymous user for token auth
		if h.anonUser != nil {
			return h.anonUser.ID
		}
		writeError(w, http.StatusInternalServerError, "anonymous user not initialized")
		return 0

	default: // AuthModeGitHub
		session, _ := h.store.Get(r, sessionName)
		userID, ok := session.Values[sessionKey].(int64)
		if !ok || userID == 0 {
			writeError(w, http.StatusUnauthorized, "authentication required — sign in with GitHub")
			return 0
		}
		return userID
	}
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
