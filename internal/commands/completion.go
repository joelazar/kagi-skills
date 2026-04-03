package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// @lat: [[cli#Local operator workflows]]
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate a shell completion script for kagi.

To load completions:

Bash:
  $ source <(kagi completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ kagi completion bash > /etc/bash_completion.d/kagi
  # macOS:
  $ kagi completion bash > $(brew --prefix)/etc/bash_completion.d/kagi

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ kagi completion zsh > "${fpath[1]}/_kagi"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ kagi completion fish | source
  # To load completions for each session, execute once:
  $ kagi completion fish > ~/.config/fish/completions/kagi.fish

PowerShell:
  PS> kagi completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, add the output to your profile:
  PS> kagi completion powershell >> $PROFILE
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
