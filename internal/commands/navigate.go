package commands

import (
	"fmt"
	"os"

	"github.com/browser-cli/internal/output"
	"github.com/spf13/cobra"
)

var navigateCmd = &cobra.Command{
	Use:   "navigate <url>",
	Short: "Navigate to a URL",
	Long: `Navigate the browser to a specified URL and wait for the page to load.

This command opens a new browser instance, navigates to the URL, and returns
page information including the final URL, title, and HTTP status.

ARGUMENTS:
  url - The URL to navigate to (must include protocol, e.g. https://)

OUTPUT:
  • url      - Final URL after navigation (may differ from input due to redirects)
  • title    - Page title
  • status   - HTTP status code (200, 404, etc.)
  • load_time - Time taken to load the page

EXAMPLES:
  browser-cli navigate https://example.com
  browser-cli navigate https://google.com
  browser-cli --output json navigate https://example.com

NOTE:
  For multi-step operations, use 'run' command instead:
  browser-cli run "navigate https://example.com; text; screenshot"`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		b, err := getBrowser()
		if err != nil {
			printError("navigate", err)
			return
		}

		result, err := b.Navigate(url, timeout)
		if err != nil {
			printError("navigate", err)
			return
		}

		printSuccess("navigate", map[string]interface{}{
			"url":       result.URL,
			"title":     result.Title,
			"status":    result.Status,
			"load_time": result.LoadTime.String(),
		})
	},
}

var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Navigate back in browser history",
	Long: `Navigate to the previous page in browser history.

This command goes back one step in the browser's history, similar to clicking
the browser's back button.

EXAMPLES:
  browser-cli run "navigate https://example.com; click a; back; text"

NOTE:
  This only works within a 'run' command sequence where browser history exists.`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := getBrowser()
		if err != nil {
			printError("back", err)
			return
		}

		if err := b.Back(timeout); err != nil {
			printError("back", err)
			return
		}

		printSuccess("back", nil)
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Navigate forward in browser history",
	Long: `Navigate to the next page in browser history.

This command goes forward one step in the browser's history, similar to clicking
the browser's forward button.

EXAMPLES:
  browser-cli run "navigate https://example.com; click a; back; forward; text"

NOTE:
  This only works within a 'run' command sequence where browser history exists.`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := getBrowser()
		if err != nil {
			printError("forward", err)
			return
		}

		if err := b.Forward(timeout); err != nil {
			printError("forward", err)
			return
		}

		printSuccess("forward", nil)
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current page",
	Long: `Reload (refresh) the current page.

This command reloads the page, similar to pressing F5 or clicking the refresh button.

EXAMPLES:
  browser-cli run "navigate https://example.com; reload; text"`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := getBrowser()
		if err != nil {
			printError("reload", err)
			return
		}

		if err := b.Reload(timeout); err != nil {
			printError("reload", err)
			return
		}

		printSuccess("reload", nil)
	},
}

func init() {
	rootCmd.AddCommand(navigateCmd)
	rootCmd.AddCommand(backCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(reloadCmd)
}

func printSuccess(cmd string, data interface{}) {
	fmt.Println(formatter.Format(output.Result{
		Command: cmd,
		Status:  "success",
		Data:    data,
	}))
}

func printError(cmd string, err error) {
	fmt.Println(formatter.Format(output.Result{
		Command: cmd,
		Status:  "error",
		Error:   err.Error(),
	}))
	os.Exit(1)
}