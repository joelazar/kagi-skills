package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Commands for managing the Kagi CLI configuration file.",
	}

	cmd.AddCommand(newConfigInitCmd(), newConfigPathCmd())

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a config file interactively",
		Long: `Create a Kagi CLI configuration file with guided prompts.

The config file is created at ~/.config/kagi/config.yaml.
Get your API key at: https://kagi.com/settings/api`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigInit(force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing config file")

	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			configDir, err := os.UserConfigDir()
			if err != nil {
				return fmt.Errorf("could not determine config directory: %w", err)
			}
			fmt.Println(filepath.Join(configDir, "kagi", "config.yaml"))
			return nil
		},
	}
}

type configFile struct {
	APIKey       string         `yaml:"api_key,omitempty"`
	SessionToken string         `yaml:"session_token,omitempty"`
	Defaults     configDefaults `yaml:"defaults,omitempty"`
}

type configDefaults struct {
	Format string       `yaml:"format,omitempty"`
	Search searchConfig `yaml:"search,omitempty"`
}

type searchConfig struct {
	Region string `yaml:"region,omitempty"`
}

func runConfigInit(force bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not determine config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "kagi", "config.yaml")

	if !force {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	config := configFile{}

	fmt.Println("Kagi CLI Configuration Setup")
	fmt.Println("============================")
	fmt.Println()

	// API key
	fmt.Println("Get your API key at: https://kagi.com/settings/api")
	fmt.Print("API key: ")

	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey != "" {
		config.APIKey = apiKey
	}

	// Session token (optional)
	fmt.Println()
	fmt.Println("Session token is optional. It enables subscriber features like")
	fmt.Println("quick answers, news, and small web feeds.")
	fmt.Print("Session token (press Enter to skip): ")

	sessionToken, _ := reader.ReadString('\n')
	sessionToken = strings.TrimSpace(sessionToken)

	if sessionToken != "" {
		config.SessionToken = sessionToken
	}

	// Default output format
	fmt.Println()
	fmt.Print("Default output format [json/compact/pretty/markdown/csv] (json): ")

	format, _ := reader.ReadString('\n')
	format = strings.TrimSpace(format)

	if format != "" && format != "json" {
		config.Defaults.Format = format
	}

	// Default search region
	fmt.Println()
	fmt.Print("Default search region (e.g., us-en, de-de; press Enter for none): ")

	region, _ := reader.ReadString('\n')
	region = strings.TrimSpace(region)

	if region != "" {
		config.Defaults.Search.Region = region
	}

	// Write config
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config) //nolint:gosec // intentionally writing API key to config file
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Config written to %s\n", configPath)
	if config.APIKey != "" {
		fmt.Println("  Run 'kagi auth check' to validate your credentials.")
	}

	return nil
}
