package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/browser-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	browserType string
	headless   bool
	timeout    time.Duration
	outputFmt   string
	sessionID   string

	// Version
	version string

	// Output formatter
	formatter *output.Formatter
)

// Execute runs the root command
func Execute(v string) {
	version = v

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "browser-cli",
	Short: "Browser automation CLI for AI agents",
	Long: `Browser-CLI is a command-line tool for browser automation, designed specifically for AI agents.

The browser server is automatically started when needed and kept running across commands.
Use --session to isolate multiple agents with independent browser instances.

KEY FEATURES:
  • Auto-managed browser server (no manual start/stop needed)
  • Multi-session support for parallel agent execution
  • Cookie persistence for login state
  • Dialog detection and handling
  • JSON output for AI parsing

USAGE:
  browser-cli navigate https://example.com    # Auto-starts server, navigates
  browser-cli click "button.submit"           # Uses existing server
  browser-cli screenshot page.png             # Takes screenshot
  browser-cli stop                            # Stops server

MULTI-AGENT:
  browser-cli --session agent-1 navigate https://site1.com
  browser-cli --session agent-2 navigate https://site2.com

OUTPUT:
  browser-cli --output json navigate https://example.com
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize formatter
		formatter = output.NewFormatter(output.Format(outputFmt))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&browserType, "browser", "b", "chromium", 
		"Browser engine: chromium (default), firefox, webkit")
	rootCmd.PersistentFlags().BoolVar(&headless, "headless", true, 
		"Run browser in headless mode (no visible window)")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, 
		"Default timeout for operations (e.g. 30s, 1m)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "markdown", 
		"Output format: json (recommended for AI), markdown (human-readable)")
	rootCmd.PersistentFlags().StringVarP(&sessionID, "session", "s", "", 
		"Session ID for isolated browser instance")
}