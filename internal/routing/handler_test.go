package routing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler(t *testing.T) (*Store, http.HandlerFunc) {
	t.Helper()
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/routing", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler(t)
	body, _ := json.Marshal(Rule{Job: "backup", Channel: "slack"})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/routing", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h(rec2, httptest.NewRequest(http.MethodGet, "/routing?job=backup", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var rule Rule
	json.NewDecoder(rec2.Body).Decode(&rule)
	if rule.Channel != "slack" {
		t.Fatalf("expected slack, got %s", rule.Channel)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/routing?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	s, h := makeHandler(t)
	_ = s.Set("backup", "slack")
	_ = s.Set("deploy", "email")
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/routing", nil))
	var rules []Rule
	json.NewDecoder(rec.Body).Decode(&rules)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	s, h := makeHandler(t)
	_ = s.Set("backup", "slack")
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/routing?job=backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected rule to be deleted")
	}
}

func TestHTTPHandler_PutInvalidJSON(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/routing", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
