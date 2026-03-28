package api_test

import (
	"testing"

	"github.com/joelazar/kagi/internal/api"
	"github.com/joelazar/kagi/internal/config"
)

func TestResolveAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		cfg     *config.Config
		want    string
		wantErr bool
	}{
		{
			name:   "from env var",
			envKey: "test-key-from-env",
			cfg:    nil,
			want:   "test-key-from-env",
		},
		{
			name: "from config",
			cfg:  &config.Config{APIKey: "test-key-from-config"},
			want: "test-key-from-config",
		},
		{
			name:   "env takes priority over config",
			envKey: "env-key",
			cfg:    &config.Config{APIKey: "config-key"},
			want:   "env-key",
		},
		{
			name:    "no key available",
			wantErr: true,
		},
		{
			name:    "nil config no env",
			cfg:     &config.Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv("KAGI_API_KEY", tt.envKey)
			} else {
				t.Setenv("KAGI_API_KEY", "")
			}

			got, err := api.ResolveAPIKey(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ResolveAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveSessionToken(t *testing.T) {
	tests := []struct {
		name     string
		envToken string
		cfg      *config.Config
		want     string
		wantErr  bool
	}{
		{
			name:     "from env var",
			envToken: "session-from-env",
			want:     "session-from-env",
		},
		{
			name: "from config",
			cfg:  &config.Config{SessionToken: "session-from-config"},
			want: "session-from-config",
		},
		{
			name:    "no token available",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envToken != "" {
				t.Setenv("KAGI_SESSION_TOKEN", tt.envToken)
			} else {
				t.Setenv("KAGI_SESSION_TOKEN", "")
			}

			got, err := api.ResolveSessionToken(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveSessionToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ResolveSessionToken() = %q, want %q", got, tt.want)
			}
		})
	}
}
