package ack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler() (*Store, http.Handler) {
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/acks", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []Ack
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty list, got %d items", len(out))
	}
}

func TestHTTPHandler_AddAndList(t *testing.T) {
	_, h := makeHandler()
	body := `{"job_name":"backup","duration":"2h","reason":"planned"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewBufferString(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/acks", nil))
	var out []Ack
	_ = json.NewDecoder(rec2.Body).Decode(&out)
	if len(out) != 1 || out[0].JobName != "backup" {
		t.Fatalf("unexpected acks: %+v", out)
	}
}

func TestHTTPHandler_AddInvalidDuration(t *testing.T) {
	_, h := makeHandler()
	body := `{"job_name":"backup","duration":"bad"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewBufferString(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_Remove(t *testing.T) {
	s, h := makeHandler()
	s.Add("backup", "", 2*60*60*1000000000)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/acks?job=backup", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if s.IsAcked("backup") {
		t.Fatal("expected ack to be removed")
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/acks", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
