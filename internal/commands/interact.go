package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var clickCmd = &cobra.Command{
	Use:   "click <selector>",
	Short: "Click an element on the page",
	Long: `Click on an element identified by a CSS selector.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector to identify the element (e.g. "#button", ".submit", "a[href]")

SELECTOR EXAMPLES:
  • "#submit"           - Element with id="submit"
  • ".btn-primary"      - Elements with class "btn-primary"
  • "button[type=submit]" - Submit buttons
  • "text=Login"        - Element containing text "Login"

EXAMPLES:
  browser-cli click "#submit-button"
  browser-cli click "text=Login"
  browser-cli click "button.submit"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("click", map[string]interface{}{"selector": args[0]})
	},
}

var clickJsCmd = &cobra.Command{
	Use:   "click-js <selector>",
	Short: "Click an element using JavaScript (bypasses visibility checks)",
	Long: `Click an element using JavaScript directly, bypassing Playwright's visibility checks.
Useful for clicking elements that are hidden, obscured, or have Vue/React event handlers.

ARGUMENTS:
  selector - CSS selector to identify the element

EXAMPLES:
  browser-cli click-js "#hidden-button"
  browser-cli click-js ".vue-tab"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("click-js", map[string]interface{}{"selector": args[0]})
	},
}

var smartClickCmd = &cobra.Command{
	Use:   "smart-click <selector>",
	Short: "Intelligently click an element, handling Web Components automatically",
	Long: `Intelligently click an element using multiple methods, automatically handling Web Components.

This command tries multiple click methods in order:
  1. Regular DOM click()
  2. Shadow DOM internal button click
  3. Internal callback functions (_onClick, _onPublish, _handleSubmit, etc.)
  4. Light DOM internal button click
  5. Custom events dispatch
  6. Framework-specific triggers (Vue/React)

ARGUMENTS:
  selector - CSS selector for the element (works with custom Web Components)

EXAMPLES:
  browser-cli smart-click "custom-button"
  browser-cli smart-click "#submit-button"
  browser-cli smart-click "[data-action=publish]"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("smart-click", map[string]interface{}{"selector": args[0]})
	},
}

var hoverCmd = &cobra.Command{
	Use:   "hover <selector>",
	Short: "Hover over an element",
	Long: `Hover over an element identified by a CSS selector.
Useful for triggering dropdown menus or hover effects.

ARGUMENTS:
  selector - CSS selector to identify the element

EXAMPLES:
  browser-cli hover "#menu-button"
  browser-cli hover ".dropdown-trigger"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("hover", map[string]interface{}{"selector": args[0]})
	},
}

var fillCmd = &cobra.Command{
	Use:   "fill <selector> <value>",
	Short: "Fill an input field with a value",
	Long: `Fill an input field with a specified value.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector for the input element
  value    - The value to fill (use quotes for values with spaces)

EXAMPLES:
  browser-cli fill "#email" "user@test.com"
  browser-cli fill "input[name=user]" "john"
  browser-cli fill "#password" "secret123"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("fill", map[string]interface{}{
			"selector": args[0],
			"value":    args[1],
		})
	},
}

var selectCmd = &cobra.Command{
	Use:   "select <selector> <value>",
	Short: "Select an option from a dropdown",
	Long: `Select an option from a dropdown/select element.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector for the select element
  value    - The value of the option to select

EXAMPLES:
  browser-cli select "#country" "US"
  browser-cli select ".language" "en"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("select", map[string]interface{}{
			"selector": args[0],
			"value":    args[1],
		})
	},
}

var typeCmd = &cobra.Command{
	Use:   "type <selector> <text>",
	Short: "Type text into an element character by character",
	Long: `Type text into an element with realistic keystroke delays.

Unlike 'fill' which sets the value directly, 'type' simulates actual typing.
The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector for the element
  text     - The text to type (use quotes for text with spaces)

FLAGS:
  --delay  - Delay between keystrokes in milliseconds (default: 50)

EXAMPLES:
  browser-cli type "#search" "hello world"
  browser-cli type "#message" "Hello!" --delay 100`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		delay, _ := cmd.Flags().GetInt("delay")
		return sendCommand("type", map[string]interface{}{
			"selector": args[0],
			"text":     args[1],
			"delay":    delay,
		})
	},
}

var pickCmd = &cobra.Command{
	Use:   "pick <x> <y>",
	Short: "Pick element at coordinates and show DOM hierarchy with detected methods",
	Long: `Pick element at screen coordinates and return detailed DOM information.

This command helps discover Web Component internals and element hierarchy for debugging.
Returns: element info, ancestor chain, detected callable methods, and Shadow DOM structure.

ARGUMENTS:
  x - X coordinate (pixels from left edge of viewport)
  y - Y coordinate (pixels from top edge of viewport)

FLAGS:
  --depth - Number of ancestor levels to traverse (default: 5)

OUTPUT:
  - target: The element at coordinates (tag, text, selector, attributes)
  - ancestors: Parent chain with children summary and detected methods
  - shadowDOM: Shadow DOM structure if present
  - suggestions: Recommended actions (e.g., "Use smart-click for Web Component")

EXAMPLES:
  browser-cli pick 500 300
  browser-cli pick 100 200 --depth=10

USE CASES:
  - Discover internal methods like _onPublish, _onClick on Web Components
  - Find the correct selector for nested elements
  - Understand Shadow DOM structure
  - Debug why click() doesn't work on custom elements`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		depth, _ := cmd.Flags().GetInt("depth")
		x, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid x coordinate: %w", err)
		}
		y, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid y coordinate: %w", err)
		}
		return sendCommand("pick", map[string]interface{}{
			"x":     x,
			"y":     y,
			"depth": depth,
		})
	},
}

func init() {
	typeCmd.Flags().Int("delay", 50, "Delay between keystrokes in ms")
	pickCmd.Flags().Int("depth", 5, "Number of ancestor levels to traverse")

	rootCmd.AddCommand(clickCmd)
	rootCmd.AddCommand(clickJsCmd)
	rootCmd.AddCommand(smartClickCmd)
	rootCmd.AddCommand(hoverCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(pickCmd)
}