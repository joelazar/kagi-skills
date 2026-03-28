package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Meta represents the metadata returned by Kagi API responses.
type Meta struct {
	ID         string   `json:"id,omitempty"`
	Node       string   `json:"node,omitempty"`
	MS         int      `json:"ms,omitempty"`
	APIBalance *float64 `json:"api_balance,omitempty"`
}

// BalanceCache represents cached API balance information.
type BalanceCache struct {
	APIBalance float64 `json:"api_balance"`
	UpdatedAt  string  `json:"updated_at"`
	Source     string  `json:"source,omitempty"`
}

// SaveBalanceCache persists the API balance from a response metadata.
func SaveBalanceCache(meta Meta, source string) error {
	if meta.APIBalance == nil {
		return nil
	}
	path, err := BalanceCachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	cached := BalanceCache{
		APIBalance: *meta.APIBalance,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
		Source:     source,
	}
	payload, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o600)
}

// LoadBalanceCache reads the cached API balance from disk.
func LoadBalanceCache() (BalanceCache, error) {
	path, err := BalanceCachePath()
	if err != nil {
		return BalanceCache{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return BalanceCache{}, err
	}
	var out BalanceCache
	if err := json.Unmarshal(b, &out); err != nil {
		return BalanceCache{}, err
	}
	return out, nil
}

// BalanceCachePath returns the filesystem path for the balance cache file.
func BalanceCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "kagi", "api_balance.json"), nil
}
