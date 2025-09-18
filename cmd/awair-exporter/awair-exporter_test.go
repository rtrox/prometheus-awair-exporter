package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	rw := httptest.NewRecorder()
	h := newHealthCheckHandler()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Errorf("/healthz returned %d, want 200", rw.Code)
	}
	if !strings.Contains(rw.Body.String(), "OK") {
		t.Errorf("/healthz body = %q, want contains 'OK'", rw.Body.String())
	}
}

func TestMetricsHandler_NoHostname(t *testing.T) {
	handler := newMetricsHandler("", false, false)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("/metrics request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/metrics returned %d, want 200", resp.StatusCode)
	}
}

func TestMetricsHandler_WithHostname(t *testing.T) {
	// This will attempt to connect to the hostname, so we expect a 502 Bad Gateway
	handler := newMetricsHandler("dummy-host", false, false)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("/metrics request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway && resp.StatusCode != http.StatusOK {
		t.Errorf("/metrics returned %d, want 200 or 502", resp.StatusCode)
	}
}

func TestProbeHandler_NoTarget(t *testing.T) {
	handler := newProbeHandler()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("/probe request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("/probe with no target returned %d, want 400", resp.StatusCode)
	}
}

func TestProbeHandler_WithTarget(t *testing.T) {
	handler := newProbeHandler()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "?target=dummy-host")
	if err != nil {
		t.Fatalf("/probe request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway && resp.StatusCode != http.StatusOK {
		t.Errorf("/probe with dummy target returned %d, want 502 or 200", resp.StatusCode)
	}
}
