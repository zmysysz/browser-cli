package commands

import (
	"time"

	"github.com/spf13/cobra"
)

var navigateCmd = &cobra.Command{
	Use:   "navigate <url>",
	Short: "Navigate to a URL",
	Long: `Navigate the browser to a specified URL.

The browser server is auto-started if not running.

ARGUMENTS:
  url - The URL to navigate to (must include protocol, e.g. https://)

FLAGS:
  --wait-for <selector>  Wait for an element to appear after navigation
  --wait-timeout <dur>   Timeout for --wait-for (default: 30s)

EXAMPLES:
  browser-cli navigate https://example.com
  browser-cli navigate https://example.com --wait-for ".dashboard"
  browser-cli --session agent-1 navigate https://site.com
  browser-cli --output json navigate https://example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := map[string]interface{}{"url": args[0]}
		if wf, _ := cmd.Flags().GetString("wait-for"); wf != "" {
			params["wait_for"] = wf
		}
		if wt, _ := cmd.Flags().GetDuration("wait-timeout"); wt > 0 {
			params["wait_timeout"] = float64(wt.Milliseconds())
		}
		return sendCommand("navigate", params)
	},
}


var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Navigate back in browser history",
	Long: `Navigate to the previous page in browser history.

The browser server is auto-started if not running.

EXAMPLES:
  browser-cli back`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("back", nil)
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Navigate forward in browser history",
	Long: `Navigate to the next page in browser history.

The browser server is auto-started if not running.

EXAMPLES:
  browser-cli forward`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("forward", nil)
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current page",
	Long: `Reload (refresh) the current page.

The browser server is auto-started if not running.

EXAMPLES:
  browser-cli reload`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("reload", nil)
	},
}

func init() {
	rootCmd.AddCommand(navigateCmd)
	rootCmd.AddCommand(backCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(reloadCmd)

	navigateCmd.Flags().String("wait-for", "", "Wait for a CSS selector to appear after navigation")
	navigateCmd.Flags().Duration("wait-timeout", 30*time.Second, "Timeout for --wait-for")
}