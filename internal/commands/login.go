package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/browser-cli/internal/browser"
)

var (
	loginStatePath string
	loginChrome    bool
	loginProxy     string
)

// loginCmd opens a browser for the user to manually log in to a website,
// then saves the storage state (cookies + localStorage) to a file.
// Subsequent sessions can reuse this state via --state <path>.
var loginCmd = &cobra.Command{
	Use:   "login <url>",
	Short: "Manually log in to a website and save the session state",
	Long: `Open a browser window for you to manually log in to a website.
After you finish logging in, press Ctrl+C to save the session state
(cookies + localStorage) to a file. Subsequent browser-cli sessions
can reuse this state via the --state flag.

This is the recommended way to handle Google login and other sites
with anti-automation detection — you log in once as a real human,
then reuse the authenticated state for automation.

EXAMPLES:
  # Log in to Google and save state to default path
  browser-cli login https://accounts.google.com

  # Log in to Google with Chrome and custom state path
  browser-cli login --chrome --state ./google-state.json https://accounts.google.com

  # Later, reuse the saved state for automation
  browser-cli --state ./google-state.json navigate https://myaccount.google.com`,
	Args: cobra.ExactArgs(1),
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	url := args[0]

	// Default state path
	if loginStatePath == "" {
		loginStatePath = browser.DefaultStatePath()
	}

	// Check if server is already running
	client := browser.NewClient("")
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		return fmt.Errorf("a browser server is already running. Stop it first with 'browser-cli stop'")
	}

	// Start server with --chrome and --headless=false (user must see the browser)
	cfg := browser.ServerConfig{
		Browser:     "chromium",
		Headless:    false, // Must be visible for manual login
		Proxy:       loginProxy,
		IdleTimeout: 0, // No auto-shutdown while user is logging in
		Chrome:      loginChrome,
		StatePath:   "", // Don't load state — we're creating a new one
	}

	fmt.Printf("Opening browser for manual login to: %s\n", url)
	fmt.Printf("State will be saved to: %s\n", loginStatePath)
	fmt.Println("Log in with your credentials, then press Ctrl+C to save and exit.")

	server, err := browser.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}

	// Start server in background
	go server.Start()

	// Wait for server to be ready
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		client = browser.NewClient("")
		resp, err = client.SendCommand(browser.Command{Action: "status"})
		if err == nil && resp.Success {
			break
		}
	}

	// Use a fresh session ID to avoid loading legacy cookies
	loginSession := "login-" + fmt.Sprintf("%d", time.Now().Unix())

	// Navigate to the login URL
	navResp, err := client.SendCommand(browser.Command{
		Action:    "navigate",
		SessionID: loginSession,
		Params:    map[string]interface{}{"url": url},
	})
	if err != nil {
		server.Stop()
		return fmt.Errorf("failed to navigate: %w", err)
	}
	if !navResp.Success {
		server.Stop()
		return fmt.Errorf("navigation failed: %s", navResp.Error)
	}

	fmt.Println("Browser is open. Complete your login, then press Ctrl+C...")

	// Wait for Ctrl+C — save state BEFORE stopping the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nSaving session state...")

	// Save storage state first (before server.Stop closes the browser)
	saveResp, err := client.SendCommand(browser.Command{
		Action:    "storage_state_save",
		SessionID: loginSession,
		Params:    map[string]interface{}{"path": loginStatePath},
	})
	if err != nil {
		// Even if save fails, try to stop cleanly
		server.Stop()
		return fmt.Errorf("failed to save state: %w", err)
	}
	if !saveResp.Success {
		server.Stop()
		return fmt.Errorf("save failed: %s", saveResp.Error)
	}

	fmt.Printf("✓ State saved to %s\n", loginStatePath)
	fmt.Printf("  Use --state %s to reuse this login in future sessions.\n", loginStatePath)

	server.Stop()
	return nil
}

// stateCmd manages saved browser states
var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Manage saved browser login states",
	Long: `Manage saved browser login states (storage state files).

Storage state files contain cookies and localStorage from a manual
login session. Use them to skip login in automated sessions.

EXAMPLES:
  # Save current session state
  browser-cli state save ./google-state.json

  # Save to default path
  browser-cli state save

  # Load state into current session
  browser-cli state load ./google-state.json`,
}

var stateSaveCmd = &cobra.Command{
	Use:   "save [path]",
	Short: "Save current session state",
	Long: `Save the current browser session's cookies and localStorage to a file.

If no path is specified, saves to the default location.

EXAMPLES:
  browser-cli state save ./google-state.json
  browser-cli state save`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		return sendCommand("storage_state_save", map[string]interface{}{
			"path": path,
		})
	},
}

var stateLoadCmd = &cobra.Command{
	Use:   "load <path>",
	Short: "Load saved state into current session",
	Long: `Load a previously saved storage state file into the current session.

This injects cookies and localStorage from the file, allowing you
to resume an authenticated session without logging in again.

EXAMPLES:
  browser-cli state load ./google-state.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("storage_state_load", map[string]interface{}{
			"path": args[0],
		})
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(stateCmd)

	loginCmd.Flags().StringVar(&loginStatePath, "state", "", "Path to save storage state (default: <data-dir>/state/default.json)")
	loginCmd.Flags().BoolVar(&loginChrome, "chrome", false, "Use system-installed Google Chrome")
	loginCmd.Flags().StringVar(&loginProxy, "proxy", "", "Proxy server URL")

	stateCmd.AddCommand(stateSaveCmd)
	stateCmd.AddCommand(stateLoadCmd)
}
