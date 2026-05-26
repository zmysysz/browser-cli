package commands

import (
	"github.com/spf13/cobra"
)

var waitCmd = &cobra.Command{
	Use:   "wait <selector>",
	Short: "Wait for an element to appear on the page",
	Long: `Wait for an element matching the selector to appear on the page.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector for the element to wait for

EXAMPLES:
  browser-cli wait ".content"
  browser-cli wait "button.submit"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("wait", map[string]interface{}{"selector": args[0]})
	},
}

var scrollCmd = &cobra.Command{
	Use:   "scroll <direction>",
	Short: "Scroll the page up or down",
	Long: `Scroll the page in a specified direction.

The browser server is auto-started if not running.

ARGUMENTS:
  direction - Scroll direction: "up" or "down"

FLAGS:
  --distance - Scroll distance in pixels (default: 300)

EXAMPLES:
  browser-cli scroll down
  browser-cli scroll up
  browser-cli scroll down --distance 500`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		distance, _ := cmd.Flags().GetInt("distance")
		return sendCommand("scroll", map[string]interface{}{
			"direction": args[0],
			"distance":  distance,
		})
	},
}

func init() {
	scrollCmd.Flags().Int("distance", 300, "Scroll distance in pixels")

	rootCmd.AddCommand(waitCmd)
	rootCmd.AddCommand(scrollCmd)
}