package ack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler() (*Store, http.Handler) {
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/acks", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result []Acknowledgement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d items", len(result))
	}
}

func TestHTTPHandler_AddAndList(t *testing.T) {
	_, h := makeHandler()
	body := `{"job":"backup","acked_by":"alice","duration":"1h"}`
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewBufferString(body)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/acks", nil))
	var result []Acknowledgement
	_ = json.NewDecoder(rr2.Body).Decode(&result)
	if len(result) != 1 || result[0].Job != "backup" || result[0].AckedBy != "alice" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestHTTPHandler_AddInvalidDuration(t *testing.T) {
	_, h := makeHandler()
	body := `{"job":"backup","acked_by":"alice","duration":"bad"}`
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/acks", bytes.NewBufferString(body)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHTTPHandler_Remove(t *testing.T) {
	s, h := makeHandler()
	s.Acknowledge("cleanup", "bob", time.Hour)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/acks/cleanup", nil))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if s.IsAcknowledged("cleanup") {
		t.Fatal("expected acknowledgement to be removed")
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPut, "/acks", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
