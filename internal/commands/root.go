package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	browserType string
	headless    bool
	timeout     time.Duration
	outputFmt   string
	sessionID   string
	proxy       string
	idleTimeout time.Duration

	// Version
	version string
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
		// Reserved for future global setup (e.g. structured logger binding).
		_, _, _ = cmd, args, outputFmt
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
	rootCmd.PersistentFlags().StringVarP(&sessionID, "session", "s", "default",
		"Session ID for isolated browser context (each session gets its own cookies/storage)")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "",
		"Proxy server URL (e.g. http://proxy.example.com:8080 or socks5://proxy:1080)")
	rootCmd.PersistentFlags().DurationVar(&idleTimeout, "idle-timeout", 1*time.Hour,
		"Auto-shutdown server after idle period (e.g. 30m, 1h, 0 to disable)")
}