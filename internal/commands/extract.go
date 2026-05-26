package commands

import (
	"github.com/spf13/cobra"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [file]",
	Short: "Take a screenshot of the current page",
	Long: `Take a screenshot of the current page and save it to a file.

The browser server is auto-started if not running.

ARGUMENTS:
  file - Optional filename for the screenshot (default: screenshot.png)

EXAMPLES:
  browser-cli screenshot
  browser-cli screenshot page.png
  browser-cli screenshot /tmp/capture.png`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "screenshot.png"
		if len(args) > 0 {
			path = args[0]
		}
		return sendCommand("screenshot", map[string]interface{}{"path": path})
	},
}

var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Extract all visible text from the page",
	Long: `Extract all visible text content from the current page.

The browser server is auto-started if not running.

OUTPUT:
  • content - The visible text content of the page

EXAMPLES:
  browser-cli text
  browser-cli --output json text`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("text", nil)
	},
}

var elementsCmd = &cobra.Command{
	Use:   "elements <selector>",
	Short: "Find and list elements matching a selector",
	Long: `Find all elements matching a CSS selector and return their information.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector to match elements

OUTPUT:
  • count  - Number of elements found
  • items  - Array of element information

SELECTOR EXAMPLES:
  • "a"              - All links
  • "input"          - All input elements
  • ".article"       - All elements with class "article"

EXAMPLES:
  browser-cli elements "a"
  browser-cli elements ".headline"
  browser-cli elements "input[type=text]"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("elements", map[string]interface{}{"selector": args[0]})
	},
}

var evalCmd = &cobra.Command{
	Use:   "eval <javascript>",
	Short: "Execute JavaScript in the browser context",
	Long: `Execute JavaScript code in the browser and return the result.

The browser server is auto-started if not running.

ARGUMENTS:
  javascript - JavaScript expression or code to execute

EXAMPLES:
  browser-cli eval "document.title"
  browser-cli eval "document.querySelector('h1').textContent"
  browser-cli eval "window.location.href"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("eval", map[string]interface{}{"script": args[0]})
	},
}

func init() {
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(textCmd)
	rootCmd.AddCommand(elementsCmd)
	rootCmd.AddCommand(evalCmd)
}