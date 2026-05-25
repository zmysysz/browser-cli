package commands

import (
	"github.com/spf13/cobra"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [file]",
	Short: "Take a screenshot of the current page",
	Long: `Take a screenshot of the current page and save it to a file.

This command captures the entire visible page (or full page if scrollable)
and saves it as a PNG image.

ARGUMENTS:
  file - Optional filename for the screenshot (default: screenshot.png)

OUTPUT:
  • path - The file path where the screenshot was saved

EXAMPLES:
  browser-cli run "navigate https://example.com; screenshot"
  browser-cli run "navigate https://example.com; screenshot page.png"
  browser-cli run "navigate https://example.com; screenshot /tmp/capture.png"

NOTE:
  Screenshots are saved in PNG format.
  Use absolute paths to save to specific directories.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "screenshot.png"
		if len(args) > 0 {
			path = args[0]
		}

		b, err := getBrowser()
		if err != nil {
			printError("screenshot", err)
			return
		}

		result, err := b.Screenshot(path)
		if err != nil {
			printError("screenshot", err)
			return
		}

		printSuccess("screenshot", map[string]interface{}{
			"path": result.Path,
		})
	},
}

var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Extract all visible text from the page",
	Long: `Extract all visible text content from the current page.

This command retrieves the text content of the page body, useful for:
  • Reading page content for analysis
  • Extracting data from simple pages
  • Checking page content after navigation

OUTPUT:
  • content - The visible text content of the page

EXAMPLES:
  browser-cli run "navigate https://example.com; text"
  browser-cli --output json run "navigate https://news.com; text"

NOTE:
  This extracts visible text only. Hidden elements and scripts are excluded.
  For structured data extraction, use 'elements' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := getBrowser()
		if err != nil {
			printError("text", err)
			return
		}

		result, err := b.Text()
		if err != nil {
			printError("text", err)
			return
		}

		printSuccess("text", map[string]interface{}{
			"content": result.Text,
		})
	},
}

var elementsCmd = &cobra.Command{
	Use:   "elements <selector>",
	Short: "Find and list elements matching a selector",
	Long: `Find all elements matching a CSS selector and return their information.

This command searches for elements and returns detailed information about each,
including tag name, text content, attributes, and visibility.

ARGUMENTS:
  selector - CSS selector to match elements

OUTPUT:
  • count  - Number of elements found
  • items  - Array of element information:
    • Tag     - HTML tag name (A, DIV, INPUT, etc.)
    • Text    - Text content of the element
    • ID      - Element ID attribute
    • Class   - Element class attribute
    • Href    - href attribute (for links)
    • Visible - Whether element is visible

SELECTOR EXAMPLES:
  • "a"              - All links
  • "a[href]"        - All links with href attribute
  • "input"          - All input elements
  • ".article"       - All elements with class "article"
  • "#main-content"  - Element with specific ID
  • "div > p"        - Paragraphs inside divs

EXAMPLES:
  browser-cli run "navigate https://example.com; elements a"
  browser-cli --output json run "navigate https://news.com; elements '.headline'"
  browser-cli run "navigate https://shop.com; elements 'input[type=text]'"`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		selector := args[0]

		b, err := getBrowser()
		if err != nil {
			printError("elements", err)
			return
		}

		elements, err := b.Elements(selector)
		if err != nil {
			printError("elements", err)
			return
		}

		printSuccess("elements", map[string]interface{}{
			"selector": selector,
			"count":    len(elements),
			"items":    elements,
		})
	},
}

var evalCmd = &cobra.Command{
	Use:   "eval <javascript>",
	Short: "Execute JavaScript in the browser context",
	Long: `Execute JavaScript code in the browser and return the result.

This command runs JavaScript in the page context, allowing you to:
  • Get page information (document.title, document.URL)
  • Extract data from the DOM
  • Modify page content
  • Call page functions

ARGUMENTS:
  javascript - JavaScript expression or code to execute

OUTPUT:
  • result - The return value of the JavaScript execution

EXAMPLES:
  browser-cli run "navigate https://example.com; eval document.title"
  browser-cli run "navigate https://example.com; eval 'document.querySelector(\"h1\").textContent'"
  browser-cli run "navigate https://example.com; eval 'window.location.href'"
  browser-cli --output json run "navigate https://example.com; eval 'JSON.stringify({title: document.title})'"

NOTE:
  Complex expressions should be quoted to avoid shell parsing issues.
  The JavaScript runs in the browser's page context, not in a sandbox.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		js := args[0]

		b, err := getBrowser()
		if err != nil {
			printError("eval", err)
			return
		}

		result, err := b.Eval(js)
		if err != nil {
			printError("eval", err)
			return
		}

		printSuccess("eval", map[string]interface{}{
			"result": result.Value,
		})
	},
}

func init() {
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(textCmd)
	rootCmd.AddCommand(elementsCmd)
	rootCmd.AddCommand(evalCmd)
}