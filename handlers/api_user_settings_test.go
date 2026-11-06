package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The API error envelope must carry the human-readable reason in "error".
// Previously it held http.StatusText, so the settings page rendered the bare
// "Bad Request" string instead of the actual problem.
func TestAPIErrorEnvelopeCarriesMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	apiError(rec, http.StatusBadRequest, "Current password is incorrect.")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (raw=%q)", err, rec.Body.String())
	}
	if body.Error != "Current password is incorrect." {
		t.Fatalf("expected error to mirror the message, got %q", body.Error)
	}
	if body.Message != "Current password is incorrect." {
		t.Fatalf("unexpected message %q", body.Message)
	}
	if body.Code != http.StatusBadRequest {
		t.Fatalf("expected code 400, got %d", body.Code)
	}
	if body.Error == http.StatusText(http.StatusBadRequest) {
		t.Fatalf("error field must not be the bare status text")
	}
}

// Without a session the endpoint must reject the request instead of reporting
// success (which previously made the UI show "Password updated successfully.").
func TestChangePasswordRejectsAnonymousRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/password",
		strings.NewReader(`{"current":"OldPassw0rd","new":"NewPassw0rd","confirm":"NewPassw0rd"}`))
	req.Header.Set("Content-Type", "application/json")

	apiChangePasswordHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (raw=%q)", err, rec.Body.String())
	}
	if body["error"] != "Authentication required." {
		t.Fatalf("expected a readable error, got %v", body["error"])
	}
	if _, hasStatus := body["status"]; hasStatus {
		t.Fatalf("anonymous requests must not report a successful status")
	}
}

// The password change flow must never answer with a body that omits a message,
// otherwise the client has nothing to display and falls back to status text.
func TestChangePasswordSuccessBodyIncludesMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	apiOK(rec, map[string]any{
		"status":  "ok",
		"ok":      true,
		"message": "Password updated successfully.",
	})
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["message"] != "Password updated successfully." {
		t.Fatalf("expected success message, got %v", body["message"])
	}
}
