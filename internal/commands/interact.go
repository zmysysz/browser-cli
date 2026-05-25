package commands

import (
	"github.com/spf13/cobra"
)

var clickCmd = &cobra.Command{
	Use:   "click <selector>",
	Short: "Click an element on the page",
	Long: `Click on an element identified by a CSS selector.

This command finds an element and performs a click action, similar to a user
clicking with their mouse.

ARGUMENTS:
  selector - CSS selector to identify the element (e.g. "#button", ".submit", "a[href]")

SELECTOR EXAMPLES:
  • "#submit"           - Element with id="submit"
  • ".btn-primary"      - Elements with class "btn-primary"
  • "button[type=submit]" - Submit buttons
  • "a[href='/login']"  - Link with specific href
  • "div > p:first-child" - First paragraph in a div

EXAMPLES:
  browser-cli run "navigate https://example.com; click '#submit-button'"
  browser-cli run "navigate https://google.com; click 'input[type=submit]'"

NOTE:
  For single commands, the browser closes after execution.
  Use 'run' command for continuous operations.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		selector := args[0]

		b, err := getBrowser()
		if err != nil {
			printError("click", err)
			return
		}

		if err := b.Click(selector, timeout); err != nil {
			printError("click", err)
			return
		}

		printSuccess("click", map[string]interface{}{
			"selector": selector,
		})
	},
}

var fillCmd = &cobra.Command{
	Use:   "fill <selector> <value>",
	Short: "Fill an input field with a value",
	Long: `Fill an input field with a specified value.

This command clears the input field and sets the value, similar to a user
typing into a form field.

ARGUMENTS:
  selector - CSS selector for the input element
  value    - The value to fill (use quotes for values with spaces)

SELECTOR EXAMPLES:
  • "#email"            - Input with id="email"
  • "input[name=user]"  - Input with name attribute
  • ".search-box"       - Input with class "search-box"
  • "textarea"          - Any textarea element

VALUE EXAMPLES:
  • 'hello'             - Simple value
  • 'hello world'       - Quoted value with spaces
  • "user@example.com"  - Email address

EXAMPLES:
  browser-cli run "navigate https://login.com; fill '#email' 'user@test.com'; fill '#password' 'secret'"
  browser-cli run "navigate https://google.com; fill '#search' 'hello world'"

NOTE:
  Use single or double quotes for values containing spaces.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		selector := args[0]
		value := args[1]

		b, err := getBrowser()
		if err != nil {
			printError("fill", err)
			return
		}

		if err := b.Fill(selector, value, timeout); err != nil {
			printError("fill", err)
			return
		}

		printSuccess("fill", map[string]interface{}{
			"selector": selector,
			"value":    value,
		})
	},
}

var selectCmd = &cobra.Command{
	Use:   "select <selector> <value>",
	Short: "Select an option from a dropdown",
	Long: `Select an option from a dropdown/select element.

This command selects an option from a <select> dropdown element by its value.

ARGUMENTS:
  selector - CSS selector for the select element
  value    - The value of the option to select

EXAMPLES:
  browser-cli run "navigate https://form.com; select '#country' 'US'"
  browser-cli run "navigate https://settings.com; select '.language' 'en'"

NOTE:
  The value must match one of the option values in the dropdown,
  not the visible text of the option.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		selector := args[0]
		value := args[1]

		b, err := getBrowser()
		if err != nil {
			printError("select", err)
			return
		}

		if err := b.Select(selector, value, timeout); err != nil {
			printError("select", err)
			return
		}

		printSuccess("select", map[string]interface{}{
			"selector": selector,
			"value":    value,
		})
	},
}

var typeCmd = &cobra.Command{
	Use:   "type <selector> <text>",
	Short: "Type text into an element character by character",
	Long: `Type text into an element with realistic keystroke delays.

Unlike 'fill' which sets the value directly, 'type' simulates actual typing
with delays between keystrokes. This is useful for:
  • Inputs that react to keystroke events
  • Search boxes with autocomplete
  • Testing keyboard interactions

ARGUMENTS:
  selector - CSS selector for the element
  text     - The text to type (use quotes for text with spaces)

FLAGS:
  --delay  - Delay between keystrokes in milliseconds (default: 50)

EXAMPLES:
  browser-cli run "navigate https://google.com; type '#search' 'hello world'"
  browser-cli run "navigate https://chat.com; type '#message' 'Hello!' --delay 100"`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		selector := args[0]
		text := args[1]

		b, err := getBrowser()
		if err != nil {
			printError("type", err)
			return
		}

		delay, _ := cmd.Flags().GetInt("delay")
		if err := b.Type(selector, text, delay, timeout); err != nil {
			printError("type", err)
			return
		}

		printSuccess("type", map[string]interface{}{
			"selector": selector,
			"text":     text,
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