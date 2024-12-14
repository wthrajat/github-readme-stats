package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusUpJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/status/up?type=json", nil)
	w := httptest.NewRecorder()
	HandleStatusUp(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("bad content type %q", ct)
	}
}

func TestTopLangsBadLayout(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/top-langs?username=x&layout=bogus", nil)
	w := httptest.NewRecorder()
	HandleTopLangs(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Incorrect layout") {
		t.Error("should reject bad layout")
	}
}

func TestStatsMissingUsername(t *testing.T) {
	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()
	HandleStats(w, req)
	if !strings.Contains(w.Body.String(), "Missing params") {
		t.Errorf("should report missing username, got: %.200s", w.Body.String())
	}
}

func TestGistMissingID(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/gist", nil)
	w := httptest.NewRecorder()
	HandleGist(w, req)
	if !strings.Contains(w.Body.String(), "Missing params") {
		t.Errorf("should report missing id, got: %.200s", w.Body.String())
	}
}
