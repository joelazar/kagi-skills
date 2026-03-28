package api_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/joelazar/kagi/internal/api"
)

func TestNewHTTPClient(t *testing.T) {
	client := api.NewHTTPClient(5 * time.Second)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", client.Timeout)
	}
}

func TestDoAPIRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":"ok"}`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(5 * time.Second)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)

	body, err := api.DoAPIRequest(client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"data":"ok"}` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestDoAPIRequest_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":[{"code":401,"msg":"Unauthorized"}]}`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(5 * time.Second)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)

	_, err := api.DoAPIRequest(client, req)
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"0.0.0.0", true},
		{"100.64.0.1", true},
		{"100.127.255.255", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"100.63.255.255", false},
		{"100.128.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := api.IsBlockedIP(ip); got != tt.blocked {
				t.Errorf("IsBlockedIP(%s) = %v, want %v", tt.ip, got, tt.blocked)
			}
		})
	}
}

func TestIsBlockedIP_Nil(t *testing.T) {
	if !api.IsBlockedIP(nil) {
		t.Error("expected nil IP to be blocked")
	}
}

func TestValidateRemoteFetchURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://example.com/path", false},
		{"ftp://example.com", true},
		{"file:///etc/passwd", true},
		{"not-a-url", true},
		{"https://127.0.0.1", true},
		{"https://10.0.0.1/foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			_, err := api.ValidateRemoteFetchURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRemoteFetchURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
