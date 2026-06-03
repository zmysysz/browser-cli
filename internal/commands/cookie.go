package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cookieCmd = &cobra.Command{
	Use:   "cookie",
	Short: "Manage browser cookies",
	Long: `Manage saved cookies for maintaining login state across sessions.

Cookies are automatically saved when a session closes and loaded when it starts.
Each session has its own isolated cookie storage.
Use this command to inspect or clear saved cookies.

EXAMPLES:
  # List all saved cookies for current session
  browser-cli cookie list
  
  # Clear cookies for a specific domain
  browser-cli cookie clear example.com
  
  # Clear all cookies for current session
  browser-cli cookie clear --all`,
}

var cookieListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved cookies",
	Long: `List all saved cookies grouped by domain for the current session.

OUTPUT:
  Returns a list of domains with their cookie count.
  Use --output json for structured data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("cookie_list", nil)
	},
}

var cookieClearCmd = &cobra.Command{
	Use:   "clear [domain]",
	Short: "Clear saved cookies",
	Long: `Clear saved cookies for a specific domain or all domains in the current session.

ARGUMENTS:
  domain - Optional. The domain to clear cookies for.
           If not provided, use --all to clear all cookies.

FLAGS:
  --all - Clear all saved cookies for the current session.

EXAMPLES:
  # Clear cookies for example.com
  browser-cli cookie clear example.com
  
  # Clear all cookies for current session
  browser-cli cookie clear --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clearAll, _ := cmd.Flags().GetBool("all")

		var domain string
		if len(args) > 0 {
			domain = args[0]
		}

		if domain == "" && !clearAll {
			return fmt.Errorf("specify a domain or use --all")
		}

		params := map[string]interface{}{
			"all": clearAll,
		}
		if domain != "" {
			params["domain"] = domain
		}

		return sendCommand("cookie_clear", params)
	},
}

func init() {
	cookieClearCmd.Flags().Bool("all", false, "Clear all saved cookies for the current session")

	cookieCmd.AddCommand(cookieListCmd)
	cookieCmd.AddCommand(cookieClearCmd)
	rootCmd.AddCommand(cookieCmd)
}
