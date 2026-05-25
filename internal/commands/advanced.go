package commands

import (
	"github.com/spf13/cobra"
)

var waitCmd = &cobra.Command{
	Use:   "wait <selector>",
	Short: "Wait for an element to appear on the page",
	Long: `Wait for an element matching the selector to appear on the page.

This command pauses execution until the specified element becomes visible,
useful for:
  • Waiting for dynamic content to load
  • Waiting for navigation to complete
  • Waiting for AJAX responses

ARGUMENTS:
  selector - CSS selector for the element to wait for

EXAMPLES:
  browser-cli run "navigate https://example.com; click '#load-more'; wait '.new-content'; text"
  browser-cli run "navigate https://search.com; fill '#q' 'hello'; click '#search'; wait '.results'"

NOTE:
  The command will timeout if the element doesn't appear within the default timeout.
  Use --timeout flag to adjust the wait duration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		selector := args[0]

		b, err := getBrowser()
		if err != nil {
			printError("wait", err)
			return
		}

		if err := b.Wait(selector, timeout); err != nil {
			printError("wait", err)
			return
		}

		printSuccess("wait", map[string]interface{}{
			"selector": selector,
		})
	},
}

var scrollCmd = &cobra.Command{
	Use:   "scroll <direction>",
	Short: "Scroll the page up or down",
	Long: `Scroll the page in a specified direction.

This command scrolls the browser viewport, useful for:
  • Loading more content on infinite-scroll pages
  • Viewing content below the visible area
  • Triggering lazy-loaded elements

ARGUMENTS:
  direction - Scroll direction: "up" or "down"

FLAGS:
  --distance - Scroll distance in pixels (default: 300)

EXAMPLES:
  browser-cli run "navigate https://example.com; scroll down; screenshot"
  browser-cli run "navigate https://news.com; scroll down; scroll down; text"
  browser-cli run "navigate https://gallery.com; scroll down --distance 500"`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		direction := args[0]
		distance, _ := cmd.Flags().GetInt("distance")

		b, err := getBrowser()
		if err != nil {
			printError("scroll", err)
			return
		}

		if err := b.Scroll(direction, distance); err != nil {
			printError("scroll", err)
			return
		}

		printSuccess("scroll", map[string]interface{}{
			"direction": direction,
			"distance":  distance,
		})
	},
}

func init() {
	scrollCmd.Flags().Int("distance", 300, "Scroll distance in pixels")

	rootCmd.AddCommand(waitCmd)
	rootCmd.AddCommand(scrollCmd)
}