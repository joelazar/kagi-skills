package api

import (
	"errors"
	"os"
	"strings"

	"github.com/joelazar/kagi/internal/config"
)

// ResolveAPIKey returns the Kagi API key from environment or config.
// Priority: KAGI_API_KEY env var > config file api_key.
func ResolveAPIKey(cfg *config.Config) (string, error) {
	if key := strings.TrimSpace(os.Getenv("KAGI_API_KEY")); key != "" {
		return key, nil
	}

	if cfg != nil && cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	return "", errors.New("KAGI_API_KEY environment variable is required (https://kagi.com/settings/api)")
}

// ResolveSessionToken returns the Kagi session token from environment or config.
// Priority: KAGI_SESSION_TOKEN env var > config file session_token.
func ResolveSessionToken(cfg *config.Config) (string, error) {
	if token := strings.TrimSpace(os.Getenv("KAGI_SESSION_TOKEN")); token != "" {
		return token, nil
	}

	if cfg != nil && cfg.SessionToken != "" {
		return cfg.SessionToken, nil
	}

	return "", errors.New("KAGI_SESSION_TOKEN environment variable is required for subscriber features")
}
