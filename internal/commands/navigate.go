package commands

import (
	"github.com/spf13/cobra"
)

var navigateCmd = &cobra.Command{
	Use:   "navigate <url>",
	Short: "Navigate to a URL",
	Long: `Navigate the browser to a specified URL.

The browser server is auto-started if not running.

ARGUMENTS:
  url - The URL to navigate to (must include protocol, e.g. https://)

EXAMPLES:
  browser-cli navigate https://example.com
  browser-cli --session agent-1 navigate https://site.com
  browser-cli --output json navigate https://example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("navigate", map[string]interface{}{"url": args[0]})
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
}