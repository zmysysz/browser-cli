package commands

import (
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

func init() {
	typeCmd.Flags().Int("delay", 50, "Delay between keystrokes in ms")

	rootCmd.AddCommand(clickCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(selectCmd)
	rootCmd.AddCommand(typeCmd)
}