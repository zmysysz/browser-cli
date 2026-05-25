package commands

import (
	"fmt"

	"github.com/browser-cli/internal/browser"
	"github.com/spf13/cobra"
)

var cookieCmd = &cobra.Command{
	Use:   "cookie",
	Short: "Manage browser cookies",
	Long: `Manage saved cookies for maintaining login state across sessions.

Cookies are automatically saved when browser closes and loaded when it starts.
Use this command to inspect or clear saved cookies.

EXAMPLES:
  # List all saved cookies
  browser-cli cookie list
  
  # Clear cookies for a specific domain
  browser-cli cookie clear example.com
  
  # Clear all cookies
  browser-cli cookie clear --all`,
}

var cookieListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved cookies",
	Long: `List all saved cookies grouped by domain.

OUTPUT:
  Returns a list of domains with their cookie count.
  Use --output json for structured data.`,
	Run: func(cmd *cobra.Command, args []string) {
		infos, err := browser.GetCookieStorage().List()
		if err != nil {
			printError("cookie list", err)
			return
		}

		if len(infos) == 0 {
			printSuccess("cookie list", map[string]interface{}{
				"message": "No saved cookies",
				"domains": []browser.CookieInfo{},
			})
			return
		}

		printSuccess("cookie list", map[string]interface{}{
			"total_domains": len(infos),
			"domains":       infos,
		})
	},
}

var cookieClearCmd = &cobra.Command{
	Use:   "clear [domain]",
	Short: "Clear saved cookies",
	Long: `Clear saved cookies for a specific domain or all domains.

ARGUMENTS:
  domain - Optional. The domain to clear cookies for.
           If not provided, use --all to clear all cookies.

FLAGS:
  --all - Clear all saved cookies.

EXAMPLES:
  # Clear cookies for example.com
  browser-cli cookie clear example.com
  
  # Clear all cookies
  browser-cli cookie clear --all`,
	Run: func(cmd *cobra.Command, args []string) {
		clearAll, _ := cmd.Flags().GetBool("all")

		var domain string
		if len(args) > 0 {
			domain = args[0]
		}

		if domain == "" && !clearAll {
			printError("cookie clear", fmt.Errorf("specify a domain or use --all"))
			return
		}

		err := browser.GetCookieStorage().Clear(domain)
		if err != nil {
			printError("cookie clear", err)
			return
		}

		if clearAll {
			printSuccess("cookie clear", map[string]interface{}{
				"message": "All cookies cleared",
			})
		} else {
			printSuccess("cookie clear", map[string]interface{}{
				"message": fmt.Sprintf("Cookies cleared for %s", domain),
				"domain":  domain,
			})
		}
	},
}

func init() {
	cookieClearCmd.Flags().Bool("all", false, "Clear all saved cookies")

	cookieCmd.AddCommand(cookieListCmd)
	cookieCmd.AddCommand(cookieClearCmd)
	rootCmd.AddCommand(cookieCmd)
}