package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/browser-cli/internal/browser"
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

	// Browser instance
	browserInstance *browser.Browser

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

It provides a clean CLI interface for all common browser operations with structured output
(JSON or Markdown) that is easy for AI models to parse and understand.

KEY FEATURES:
  • Navigate, click, fill forms, take screenshots
  • Extract page text and find elements
  • Execute JavaScript in browser context
  • Continuous operations with 'run' command (recommended for AI)
  • JSON output for programmatic parsing

RECOMMENDED USAGE FOR AI:
  Use the 'run' command with JSON output for multi-step operations:
  
    browser-cli --output json run "navigate https://example.com; elements a; text"

  This keeps the browser alive across all operations and returns structured results.

BROWSER SUPPORT:
  • chromium (default) - Best compatibility
  • firefox - Alternative browser
  • webkit - Safari-like engine

EXAMPLES:
  # Single operation
  browser-cli navigate https://example.com
  
  # Continuous operations (recommended)
  browser-cli run "navigate https://example.com; click a; text"
  
  # JSON output for AI parsing
  browser-cli --output json run "navigate https://example.com; elements 'a[href]'"
  
  # Use Firefox browser
  browser-cli --browser firefox navigate https://example.com`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize formatter
		formatter = output.NewFormatter(output.Format(outputFmt))
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Cleanup browser
		if browserInstance != nil {
			browserInstance.Close()
		}
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
		"Session ID for persistent browser (not implemented yet)")
}

// getBrowser returns or creates browser instance
func getBrowser() (*browser.Browser, error) {
	if browserInstance != nil {
		return browserInstance, nil
	}

	var err error
	browserInstance, err = browser.New(browser.Config{
		Browser:  browserType,
		Headless: headless,
		Timeout:  timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create browser: %w", err)
	}

	return browserInstance, nil
}