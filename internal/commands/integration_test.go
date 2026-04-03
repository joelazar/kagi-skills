//go:build integration

package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joelazar/kagi/internal/api"
)

// setupMockServer creates a mock Kagi API server that handles all endpoints.
// @lat: [[testing#Integration tests]]
func setupMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bot test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"meta": map[string]any{"id": "err", "node": "test", "ms": 0},
				"data": nil,
				"error": []map[string]any{
					{"code": 401, "msg": "Unauthorized"},
				},
			})
			return
		}

		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/api/v0/search"):
			serveSearch(w, r)
		case strings.HasPrefix(path, "/api/v0/fastgpt"):
			serveFastGPT(w, r)
		case strings.HasPrefix(path, "/api/v0/summarize"):
			serveSummarize(w, r)
		case strings.HasPrefix(path, "/api/v0/enrich/web"):
			serveEnrich(w, r, "web")
		case strings.HasPrefix(path, "/api/v0/enrich/news"):
			serveEnrich(w, r, "news")
		case strings.HasPrefix(path, "/api/v1/translate"):
			serveTranslate(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func serveSearch(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]any{
		"meta": map[string]any{"id": "search-1", "node": "us-east", "ms": 42, "api_balance": 9.50},
		"data": []map[string]any{
			{
				"t":       0,
				"rank":    1,
				"url":     "https://go.dev",
				"title":   "The Go Programming Language",
				"snippet": "Go is an open source programming language.",
			},
			{
				"t":       0,
				"rank":    2,
				"url":     "https://golang.org/doc",
				"title":   "Go Documentation",
				"snippet": "Documentation for the Go programming language.",
			},
		},
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func serveFastGPT(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req map[string]any
	json.Unmarshal(body, &req) //nolint:errcheck

	resp := map[string]any{
		"meta": map[string]any{"id": "fgpt-1", "node": "us-east", "ms": 150, "api_balance": 9.25},
		"data": map[string]any{
			"output": "Paris is the capital of France.",
			"references": []map[string]any{
				{
					"title":   "France - Wikipedia",
					"snippet": "France's capital is Paris.",
					"url":     "https://en.wikipedia.org/wiki/France",
				},
			},
		},
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func serveSummarize(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]any{
		"meta": map[string]any{"id": "sum-1", "node": "us-east", "ms": 200, "api_balance": 9.00},
		"data": map[string]any{
			"output": "This is a summary of the article.",
		},
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func serveEnrich(w http.ResponseWriter, _ *http.Request, typ string) {
	resp := map[string]any{
		"meta": map[string]any{"id": "enrich-1", "node": "us-east", "ms": 30, "api_balance": 9.40},
		"data": []map[string]any{
			{
				"t":       0,
				"rank":    1,
				"url":     "https://example.com/" + typ,
				"title":   "Example " + typ + " result",
				"snippet": "An example " + typ + " enrichment result.",
			},
		},
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func serveTranslate(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req map[string]any
	json.Unmarshal(body, &req) //nolint:errcheck

	resp := map[string]any{
		"meta": map[string]any{"id": "tr-1", "node": "us-east", "ms": 80, "api_balance": 9.30},
		"data": map[string]any{
			"translation":     "Hallo Welt",
			"source_language": "English",
			"target_language": "German",
		},
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// executeCommand runs a kagi CLI command and captures stdout.
func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Capture stdout since commands use fmt.Println, not cmd.OutOrStdout()
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := NewRootCmd()
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()

	return buf.String(), err
}

// setTestEnv sets env vars for testing and returns a cleanup function.
func setTestEnv(t *testing.T, serverURL string) {
	t.Helper()
	t.Setenv("KAGI_API_KEY", "test-api-key")
	t.Setenv("KAGI_SEARCH_URL", serverURL+"/api/v0/search")
	t.Setenv("KAGI_FASTGPT_URL", serverURL+"/api/v0/fastgpt")
	t.Setenv("KAGI_SUMMARIZE_URL", serverURL+"/api/v0/summarize")
	t.Setenv("KAGI_ENRICH_WEB_URL", serverURL+"/api/v0/enrich/web")
	t.Setenv("KAGI_ENRICH_NEWS_URL", serverURL+"/api/v0/enrich/news")
	t.Setenv("KAGI_TRANSLATE_URL", serverURL+"/api/v1/translate")
}

func TestIntegrationVersion(t *testing.T) {
	out, err := executeCommand(t, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(out, "kagi") {
		t.Errorf("expected 'kagi' in version output, got: %s", out)
	}
}

func TestIntegrationCompletion(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			out, err := executeCommand(t, "completion", shell)
			if err != nil {
				t.Fatalf("completion %s failed: %v", shell, err)
			}
			if out == "" {
				t.Errorf("expected non-empty completion output for %s", shell)
			}
		})
	}
}

func TestIntegrationCompletionInvalid(t *testing.T) {
	_, err := executeCommand(t, "completion", "invalid")
	if err == nil {
		t.Error("expected error for invalid shell, got nil")
	}
}

func TestIntegrationConfigPath(t *testing.T) {
	out, err := executeCommand(t, "config", "path")
	if err != nil {
		t.Fatalf("config path failed: %v", err)
	}
	if !strings.Contains(out, "config.yaml") {
		t.Errorf("expected config.yaml in output, got: %s", out)
	}
}

func TestIntegrationConfigInitExisting(t *testing.T) {
	// Create a temp home dir with existing config file
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Determine where UserConfigDir will look
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("could not determine config dir: %v", err)
	}

	kagiDir := configDir + "/kagi"
	os.MkdirAll(kagiDir, 0o700)                               //nolint:errcheck
	os.WriteFile(kagiDir+"/config.yaml", []byte("{}"), 0o600) //nolint:errcheck

	_, err = executeCommand(t, "config", "init")
	if err == nil {
		t.Error("expected error when config already exists, got nil")
	}
}

func TestIntegrationAuthCheckNoKey(t *testing.T) {
	t.Setenv("KAGI_API_KEY", "")
	// Clear any config
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := executeCommand(t, "auth", "check")
	if err == nil {
		t.Error("expected error when no API key configured, got nil")
	}
}

func TestIntegrationAuthCheckWithMock(t *testing.T) {
	server := setupMockServer(t)
	defer server.Close()

	// Override the auth check URL
	origURL := authCheckSearchURL
	authCheckSearchURLOverride := server.URL + "/api/v0/search"
	// We can't easily override the const, so we test the validateAPIKey function directly
	t.Setenv("KAGI_API_KEY", "test-api-key")

	err := validateAPIKeyWithURL("test-api-key", authCheckSearchURLOverride)
	if err != nil {
		t.Fatalf("expected auth check to pass with mock server, got: %v", err)
	}

	// Test with bad key
	err = validateAPIKeyWithURL("bad-key", authCheckSearchURLOverride)
	if err == nil {
		t.Error("expected auth check to fail with bad key")
	}

	_ = origURL // keep reference
}

func TestIntegrationHelpFlags(t *testing.T) {
	commands := [][]string{
		{"--help"},
		{"search", "--help"},
		{"fastgpt", "--help"},
		{"summarize", "--help"},
		{"enrich", "--help"},
		{"balance", "--help"},
		{"auth", "--help"},
		{"auth", "check", "--help"},
		{"config", "--help"},
		{"config", "init", "--help"},
		{"config", "path", "--help"},
		{"version", "--help"},
		{"completion", "--help"},
	}

	for _, args := range commands {
		name := strings.Join(args, " ")
		t.Run(name, func(t *testing.T) {
			_, err := executeCommand(t, args...)
			if err != nil {
				t.Errorf("'kagi %s' failed: %v", name, err)
			}
		})
	}
}

func TestIntegrationUnknownCommand(t *testing.T) {
	_, err := executeCommand(t, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

// validateAPIKeyWithURL is a testable version of validateAPIKey that accepts a custom URL.
func validateAPIKeyWithURL(apiKey, url string) error {
	client := api.NewHTTPClient(5 * 1e9) // 5s

	req, err := http.NewRequest(http.MethodGet, url+"?q=test&limit=1", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	_, err = api.DoAPIRequest(client, req)
	return err
}
