package commands

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/joelazar/kagi/internal/api"
	"github.com/spf13/cobra"
)

const authCheckSearchURL = "https://kagi.com/api/v0/search"

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing and validating Kagi API credentials.",
	}

	cmd.AddCommand(newAuthCheckCmd())

	return cmd
}

func newAuthCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Validate API key configuration",
		Long: `Validate that your Kagi API key is configured and working.
Also reports whether a Kagi session token is configured for subscriber features.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runAuthCheck()
		},
	}
}

func runAuthCheck() error {
	// Check API key
	apiKey, err := api.ResolveAPIKey(cfg)
	if err != nil {
		fmt.Println("✗ API key: not configured")
		fmt.Println("  Set KAGI_API_KEY or add api_key to ~/.config/kagi/config.yaml")
		return err
	}

	fmt.Printf("✓ API key: found (%s…)\n", maskKey(apiKey))

	// Validate API key with a minimal search request
	if err := validateAPIKey(apiKey); err != nil {
		fmt.Printf("✗ API key: invalid — %v\n", err)
		return fmt.Errorf("API key validation failed: %w", err)
	}

	fmt.Println("✓ API key: valid")

	// Check session token (optional)
	sessionToken, err := api.ResolveSessionToken(cfg)
	if err != nil {
		fmt.Println("- Session token: not configured (optional, needed for subscriber features)")
	} else {
		fmt.Printf("✓ Session token: configured (%s…)\n", maskKey(sessionToken))
	}

	return nil
}

// validateAPIKey makes a minimal API call to verify the key works.
func validateAPIKey(apiKey string) error {
	client := api.NewHTTPClient(10 * time.Second)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		authCheckSearchURL+"?q=test&limit=1",
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+apiKey)
	req.Header.Set("User-Agent", api.DefaultUserAgent)

	_, err = api.DoAPIRequest(client, req)
	return err
}

// maskKey shows the first 4 and last 2 characters of a key.
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "…" + key[len(key)-2:]
}
