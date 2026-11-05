package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"platform/db"
)

// ── GET /api/user/api-keys ────────────────────────────────────

type apiKeyItem struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
}

func apiUserAPIKeysListHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	keys, err := db.ListUserPersonalAPIKeys(ctx, Pools.Admin, user.ID)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to load API keys.")
		return
	}

	items := make([]apiKeyItem, 0, len(keys))
	for _, k := range keys {
		item := apiKeyItem{
			ID:        k.ID,
			Name:      k.Name,
			CreatedAt: k.CreatedAt.UTC().Format(time.RFC3339),
		}
		if k.LastUsedAt != nil {
			s := k.LastUsedAt.UTC().Format(time.RFC3339)
			item.LastUsedAt = &s
		}
		items = append(items, item)
	}
	apiOK(w, map[string]any{"keys": items})
}

// ── POST /api/user/api-keys ───────────────────────────────────

func apiUserAPIKeysCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	if !apiKeyCreationRateLimiter.allow(fmt.Sprintf("%d", user.ID)) {
		w.Header().Set("Retry-After", "3600")
		apiError(w, http.StatusTooManyRequests, "Too many API key creation attempts. Try again in 1 hour.")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		apiError(w, http.StatusBadRequest, "A key name is required.")
		return
	}
	name := strings.TrimSpace(body.Name)
	if len(name) > 100 {
		apiError(w, http.StatusBadRequest, "Key name must not exceed 100 characters.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	const maxKeysPerUser = 10
	count, err := db.CountPersonalAPIKeys(ctx, Pools.Admin, user.ID)
	if err == nil && count >= maxKeysPerUser {
		apiError(w, http.StatusUnprocessableEntity, "API key limit reached. Revoke an existing key before creating a new one.")
		return
	}

	if exists, err := db.PersonalAPIKeyNameExists(ctx, Pools.Admin, user.ID, name); err == nil && exists {
		apiError(w, http.StatusUnprocessableEntity, "An active API key with that name already exists.")
		return
	}

	// Generate a 32-byte random key, encode as hex with a prefix.
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to generate key.")
		return
	}
	plainKey := "plat_" + hex.EncodeToString(raw[:])

	// Store only the SHA-256 hash.
	sum := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(sum[:])

	k, err := db.CreatePersonalAPIKey(ctx, Pools.Admin, user.ID, name, keyHash)
	if err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to create API key.")
		return
	}

	go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role), "api_key.created", "info", map[string]any{"name": name, "key_id": k.ID}, "", r.Header.Get("X-Request-ID"), clientIP(r))

	// Return the plaintext key once — it cannot be retrieved again.
	apiOK(w, map[string]any{
		"id":         k.ID,
		"name":       k.Name,
		"key":        plainKey,
		"created_at": k.CreatedAt.UTC().Format(time.RFC3339),
		"note":       "Copy this key now — it will not be shown again.",
	})
}

// ── DELETE /api/user/api-keys/{id} ───────────────────────────

func apiUserAPIKeyRevokeHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		apiError(w, http.StatusUnauthorized, "Authentication required.")
		return
	}
	if Pools == nil || user.ID == 0 {
		apiError(w, http.StatusServiceUnavailable, "Database not available.")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	rawID := parts[len(parts)-1]
	keyID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || keyID <= 0 {
		apiError(w, http.StatusBadRequest, "Invalid key ID.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := db.RevokePersonalAPIKey(ctx, Pools.Admin, user.ID, keyID); err != nil {
		apiError(w, http.StatusInternalServerError, "Failed to revoke key.")
		return
	}

	go db.InsertAuditEvent(context.Background(), Pools.Admin, user.ID, string(user.Role), "api_key.revoked", "info", map[string]any{"key_id": keyID}, "", r.Header.Get("X-Request-ID"), clientIP(r))

	apiOK(w, map[string]string{"status": "revoked"})
}
