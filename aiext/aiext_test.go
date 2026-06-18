package aiext

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotConfigured(t *testing.T) {
	c := &Client{http: http.DefaultClient}
	if c.Configured() {
		t.Fatal("expected not configured")
	}
	_, err := c.Process(context.Background(), CapabilityUpscale, []byte("x"), "image/png", nil)
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestProcessImageCapability(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/upscale" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["image"]; !ok {
			t.Error("missing image in request")
		}
		out := base64.StdEncoding.EncodeToString([]byte("processed"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"image": out, "mime_type": "image/png"})
	}))
	defer srv.Close()

	c := &Client{endpoint: srv.URL, http: srv.Client()}
	if !c.Configured() {
		t.Fatal("expected configured")
	}
	res, err := c.Process(context.Background(), CapabilityUpscale, []byte("raw"), "image/png", map[string]any{"scale": 2})
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if string(res.Image) != "processed" {
		t.Fatalf("unexpected image: %q", res.Image)
	}
}

func TestProcessTagCapability(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tags": []map[string]any{{"name": "cat", "score": 0.9}},
		})
	}))
	defer srv.Close()

	c := &Client{endpoint: srv.URL, http: srv.Client()}
	res, err := c.Process(context.Background(), CapabilityTag, []byte("raw"), "image/png", nil)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if len(res.Tags) != 1 || res.Tags[0].Name != "cat" {
		t.Fatalf("unexpected tags: %+v", res.Tags)
	}
}

func TestUpstreamErrorPropagated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "boom"})
	}))
	defer srv.Close()

	c := &Client{endpoint: srv.URL, http: srv.Client()}
	_, err := c.Process(context.Background(), CapabilityUpscale, []byte("raw"), "image/png", nil)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected upstream error 'boom', got %v", err)
	}
}
