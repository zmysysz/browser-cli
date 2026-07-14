package commands

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/browser-cli/internal/browser"
)




var (
	serverBrowser     string
	serverHeadless    bool
	serverSocket      string
	serverIdleTimeout time.Duration
	serverCDPEndpoint string
	serverChrome      bool
	serverStatePath   string
	serverDataDir     string
	serverLogLevel    string
	serverLogFormat   string
)


// serverCmd represents the server command (foreground, for manual use)
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a persistent browser server (foreground)",
	Long: `Start a persistent browser server in foreground mode.

The server manages a single browser instance with multiple isolated contexts.
Each --session gets its own BrowserContext with independent cookies/storage.

This is for manual use when you want to see server logs.
For normal use, the server is auto-started when needed.

EXAMPLES:
  browser-cli server --headless=false
`,
	RunE: runServer,
}

// serverStartCmd starts server in background (for auto-start)
var serverStartCmd = &cobra.Command{
	Use:   "server-start",
	Short: "Start browser server in background",
	Long:  `Start browser server in background (used internally for auto-start).`,
	RunE:  runServerStart,
}

// statusCmd shows server status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check browser server status",
	Long: `Check if the browser server is running and get its status.

OUTPUT:
  Returns server status: running, session_count, sessions list

EXAMPLES:
  browser-cli status`,
	RunE: runStatus,
}

// stopCmd stops the server
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the browser server",
	Long: `Stop the browser server and save all session cookies.

The server will be auto-started again when needed.
This closes ALL sessions.

EXAMPLES:
  browser-cli stop`,
	RunE: runStop,
}

// sessionListCmd lists all active sessions
var sessionListCmd = &cobra.Command{
	Use:   "session-list",
	Short: "List all active browser sessions",
	Long: `List all active browser sessions (contexts) in the server.

OUTPUT:
  Returns a list of session IDs

EXAMPLES:
  browser-cli session-list`,
	RunE: runSessionList,
}

// sessionCloseCmd closes a specific session
var sessionCloseCmd = &cobra.Command{
	Use:   "session-close",
	Short: "Close a browser session",
	Long: `Close a specific browser session (BrowserContext) while keeping the server running.
Cookies are automatically saved before closing.

The server itself remains running and other sessions are not affected.

EXAMPLES:
  browser-cli --session agent-1 session-close`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("session_close", nil)
	},
}

// Tab commands

var tabNewCmd = &cobra.Command{
	Use:   "tab-new",
	Short: "Create new tab",
	Long: `Create a new browser tab and switch to it.

The browser server is auto-started if not running.

EXAMPLES:
  browser-cli tab-new
  browser-cli --output json tab-new`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("tab_new", nil)
	},
}

var tabSwitchCmd = &cobra.Command{
	Use:   "tab-switch <id>",
	Short: "Switch to tab",
	Long: `Switch to a specific browser tab by its ID.

The browser server is auto-started if not running.

ARGUMENTS:
  id - Tab ID to switch to (use tab-list to see available IDs)

EXAMPLES:
  browser-cli tab-switch 2
  browser-cli tab-switch 3`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tabID, err := strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil {
			return fmt.Errorf("tab-switch: %q is not a valid tab ID: %w", args[0], err)
		}
		return sendCommand("tab_switch", map[string]interface{}{"tab_id": tabID})
	},
}

var tabListCmd = &cobra.Command{
	Use:   "tab-list",
	Short: "List all tabs",
	Long: `List all open tabs in the current browser session.

The browser server is auto-started if not running.

OUTPUT:
  Returns a list of tabs with their IDs and titles.

EXAMPLES:
  browser-cli tab-list
  browser-cli --output json tab-list`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("tab_list", nil)
	},
}

var tabCloseCmd = &cobra.Command{
	Use:   "tab-close [id]",
	Short: "Close tab",
	Long: `Close a browser tab by its ID.

The browser server is auto-started if not running.

ARGUMENTS:
  id - Optional tab ID to close (default: current active tab)

EXAMPLES:
  browser-cli tab-close        # Close current tab
  browser-cli tab-close 2      # Close tab with ID 2`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var tabID int
		if len(args) > 0 {
			id, err := strconv.Atoi(strings.TrimSpace(args[0]))
			if err != nil {
				return fmt.Errorf("tab-close: %q is not a valid tab ID: %w", args[0], err)
			}
			tabID = id
		}
		return sendCommand("tab_close", map[string]interface{}{"tab_id": tabID})
	},
}

// Dialog commands

var dialogStatusCmd = &cobra.Command{
	Use:   "dialog-status",
	Short: "Check pending dialog",
	Long: `Check if there is a pending JavaScript dialog (alert, confirm, prompt, beforeunload).

The browser server is auto-started if not running.

OUTPUT:
  Returns dialog info if pending: type, message, default_value

EXAMPLES:
  browser-cli dialog-status
  browser-cli --output json dialog-status`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("dialog_status", nil)
	},
}

var dialogAcceptCmd = &cobra.Command{
	Use:   "dialog-accept [value]",
	Short: "Accept dialog",
	Long: `Accept a pending JavaScript dialog.

The browser server is auto-started if not running.

ARGUMENTS:
  value - Optional value for prompt dialogs

EXAMPLES:
  browser-cli dialog-accept           # Accept alert/confirm
  browser-cli dialog-accept "input"   # Accept prompt with value`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := ""
		if len(args) > 0 {
			value = args[0]
		}
		return sendCommand("dialog_accept", map[string]interface{}{"value": value})
	},
}

var dialogDismissCmd = &cobra.Command{
	Use:   "dialog-dismiss",
	Short: "Dismiss dialog",
	Long: `Dismiss (cancel) a pending JavaScript dialog.

The browser server is auto-started if not running.

EXAMPLES:
  browser-cli dialog-dismiss`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("dialog_dismiss", nil)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(serverStartCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(sessionListCmd)
	rootCmd.AddCommand(sessionCloseCmd)

	// Tab commands
	rootCmd.AddCommand(tabNewCmd)
	rootCmd.AddCommand(tabSwitchCmd)
	rootCmd.AddCommand(tabListCmd)
	rootCmd.AddCommand(tabCloseCmd)

	// Dialog commands
	rootCmd.AddCommand(dialogStatusCmd)
	rootCmd.AddCommand(dialogAcceptCmd)
	rootCmd.AddCommand(dialogDismissCmd)

	// Server flags
	serverCmd.Flags().StringVar(&serverBrowser, "browser", "chromium", "Browser to use")
	serverCmd.Flags().BoolVar(&serverHeadless, "headless", true, "Run in headless mode")
	serverCmd.Flags().StringVar(&serverSocket, "socket", "", "Unix socket path")
	serverCmd.Flags().DurationVar(&serverIdleTimeout, "idle-timeout", 1*time.Hour, "Auto-shutdown after idle period (e.g. 30m, 1h, 0 to disable)")
	serverCmd.Flags().StringVar(&serverCDPEndpoint, "cdp-endpoint", "", "Connect to existing browser via CDP (e.g. http://localhost:9222)")
	serverCmd.Flags().BoolVar(&serverChrome, "chrome", false, "Use system-installed Google Chrome (via Playwright channel) instead of bundled Chromium")
	serverCmd.Flags().StringVar(&serverStatePath, "state", "", "Path to storage state JSON file (cookies+localStorage) for login reuse")
	serverCmd.Flags().StringVar(&serverDataDir, "data-dir", "", "Data directory for socket, cookies, and state (default: ~/.local/share/browser-cli or $BROWSER_CLI_HOME)")
	serverCmd.Flags().StringVar(&serverLogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	serverCmd.Flags().StringVar(&serverLogFormat, "log-format", "text", "Log format: text or json")

	serverStartCmd.Flags().StringVar(&serverBrowser, "browser", "chromium", "Browser to use")
	serverStartCmd.Flags().BoolVar(&serverHeadless, "headless", true, "Run in headless mode")
	serverStartCmd.Flags().DurationVar(&serverIdleTimeout, "idle-timeout", 1*time.Hour, "Auto-shutdown after idle period")
	serverStartCmd.Flags().StringVar(&serverCDPEndpoint, "cdp-endpoint", "", "Connect to existing browser via CDP (e.g. http://localhost:9222)")
	serverStartCmd.Flags().BoolVar(&serverChrome, "chrome", false, "Use system-installed Google Chrome (via Playwright channel) instead of bundled Chromium")
	serverStartCmd.Flags().StringVar(&serverStatePath, "state", "", "Path to storage state JSON file (cookies+localStorage) for login reuse")
	serverStartCmd.Flags().StringVar(&serverDataDir, "data-dir", "", "Data directory for socket, cookies, and state (default: ~/.local/share/browser-cli or $BROWSER_CLI_HOME)")
	serverStartCmd.Flags().StringVar(&serverLogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	serverStartCmd.Flags().StringVar(&serverLogFormat, "log-format", "text", "Log format: text or json")
}


// initServerEnv initializes the data directory and structured logger.
// Shared by runServer and runServerStart so both foreground and
// background servers get the same setup.
func initServerEnv() error {
	// Prefer the server-specific flag, fall back to the global --data-dir
	dir := serverDataDir
	if dir == "" {
		dir = dataDir
	}
	if err := browser.InitDataDir(dir); err != nil {
		return err
	}

	level := slog.LevelInfo
	switch strings.ToLower(serverLogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	format := browser.LogFormatText
	if strings.ToLower(serverLogFormat) == "json" {
		format = browser.LogFormatJSON
	}
	browser.InitLogger(format, level)
	return nil
}

func runServer(cmd *cobra.Command, args []string) error {
	// Check if server already running
	client := browser.NewClient("")
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		fmt.Println("Server already running!")
		fmt.Printf("Sessions: %v, Session count: %v\n", resp.Data["sessions"], resp.Data["session_count"])
		return nil
	}

	// Initialize data dir + logger
	if err := initServerEnv(); err != nil {
		return fmt.Errorf("failed to initialize server environment: %w", err)
	}

	// Start new server
	browser.Logger.Info("starting browser server")

	cfg := browser.ServerConfig{
		Browser:     serverBrowser,
		Headless:    serverHeadless,
		SocketPath:  serverSocket,
		Proxy:       proxy,
		IdleTimeout: serverIdleTimeout,
		CDPEndpoint: serverCDPEndpoint,
		Chrome:      serverChrome,
		StatePath:   serverStatePath,
	}

	server, err := browser.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		browser.Logger.Info("shutting down server (signal received)")
		server.Stop()
		os.Exit(0)
	}()

	return server.Start()
}

func runServerStart(cmd *cobra.Command, args []string) error {
	// Check if server already running
	client := browser.NewClient("")
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		// Already running, just exit
		return nil
	}

	// Initialize data dir + logger
	if err := initServerEnv(); err != nil {
		return fmt.Errorf("failed to initialize server environment: %w", err)
	}

	// Start new server
	cfg := browser.ServerConfig{
		Browser:     serverBrowser,
		Headless:    serverHeadless,
		SocketPath:  serverSocket,
		Proxy:       proxy,
		IdleTimeout: serverIdleTimeout,
		CDPEndpoint: serverCDPEndpoint,
		Chrome:      serverChrome,
		StatePath:   serverStatePath,
	}

	server, err := browser.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		server.Stop()
		os.Exit(0)
	}()

	return server.Start()
}


func runStatus(cmd *cobra.Command, args []string) error {
	client := browser.NewClient("")
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err != nil {
		return fmt.Errorf("server not running: %w", err)
	}

	return printResult("status", resp.Success, resp.Data, resp.Error, outputFmt)
}

func runStop(cmd *cobra.Command, args []string) error {
	client := browser.NewClient("")
	resp, err := client.SendCommand(browser.Command{Action: "stop"})
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	if resp.Success {
		fmt.Println("Server stopped successfully")
	}
	return nil
}

func runSessionList(cmd *cobra.Command, args []string) error {
	client := browser.NewClient("")
	resp, err := client.SendCommand(browser.Command{Action: "session_list"})
	if err != nil {
		return fmt.Errorf("server not running: %w", err)
	}

	return printResult("session-list", resp.Success, resp.Data, resp.Error, outputFmt)
}

// ensureServer ensures a browser server is running, starts one if not
func ensureServer() (*browser.Client, error) {
	client := browser.NewClient("")

	// Check if server is already running
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		return client, nil
	}

	// Resolve the absolute path to the current executable so that
	// symlinks, PATH lookups, and `go run` don't break auto-start.
	binaryPath, err := os.Executable()
	if err != nil {
		// Fallback to os.Args[0] if os.Executable fails (rare)
		binaryPath = os.Args[0]
	}

	// Server not running, start it in background
	serverCmd := exec.Command(binaryPath, "server-start")
	// Add headless flag only when explicitly enabled
	if headless {
		serverCmd.Args = append(serverCmd.Args, "--headless")
	}
	// Pass idle timeout
	if idleTimeout > 0 {
		serverCmd.Args = append(serverCmd.Args, "--idle-timeout", idleTimeout.String())
	} else {
		serverCmd.Args = append(serverCmd.Args, "--idle-timeout", "0")
	}
	// Pass state path for login reuse
	if statePath != "" {
		serverCmd.Args = append(serverCmd.Args, "--state", statePath)
	}
	// Pass proxy
	if proxy != "" {
		serverCmd.Args = append(serverCmd.Args, "--proxy", proxy)
	}
	// Pass data-dir so the server uses the same data directory as the client
	if dataDir != "" {
		serverCmd.Args = append(serverCmd.Args, "--data-dir", dataDir)
	}
	// Pass browser type
	if browserType != "" && browserType != "chromium" {
		serverCmd.Args = append(serverCmd.Args, "--browser", browserType)
	}

	// Start server in background
	if err := serverCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to be ready
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		resp, err = client.SendCommand(browser.Command{Action: "status"})
		if err == nil && resp.Success {
			return client, nil
		}
	}

	return nil, fmt.Errorf("server failed to start within 30 seconds")
}


// sendCommand sends a command to the server and outputs the result.
//
// Returns an error when the server reports command failure, so cobra
// produces a non-zero exit code. This lets AI agents detect success/failure
// via exit code without parsing JSON.
func sendCommand(action string, params map[string]interface{}) error {
	client, err := ensureServer()
	if err != nil {
		return err
	}

	cmd := browser.Command{
		Action:    action,
		SessionID: sessionID,
		Params:    params,
	}

	// Extract tab_id from params if present
	if tabID, ok := params["tab_id"]; ok {
		if id, ok := tabID.(int); ok {
			cmd.TabID = id
		}
	}

	resp, err := client.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return printResult(action, resp.Success, resp.Data, resp.Error, outputFmt)
}
