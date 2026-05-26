package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/browser-cli/internal/browser"
)

var (
	serverBrowser  string
	serverHeadless bool
	serverSocket   string
)

// printSuccess prints a success result
func printSuccess(cmd string, data interface{}) {
	output := map[string]interface{}{
		"command": cmd,
		"status":  "success",
	}
	if data != nil {
		output["data"] = data
	}
	if sessionID != "" {
		output["session"] = sessionID
	}
	dataBytes, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(dataBytes))
}

// printError prints an error result
func printError(cmd string, err error) {
	output := map[string]interface{}{
		"command": cmd,
		"status":  "error",
		"error":   err.Error(),
	}
	if sessionID != "" {
		output["session"] = sessionID
	}
	dataBytes, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(dataBytes))
	os.Exit(1)
}

// serverCmd represents the server command (foreground, for manual use)
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a persistent browser server (foreground)",
	Long: `Start a persistent browser server in foreground mode.

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
  Returns server status: running, active_tab, tabs count

EXAMPLES:
  browser-cli status
  browser-cli --session agent-1 status`,
	RunE: runStatus,
}

// stopCmd stops the server
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the browser server",
	Long: `Stop the browser server and save cookies.

The server will be auto-started again when needed.

EXAMPLES:
  browser-cli stop
  browser-cli --session agent-1 stop`,
	RunE: runStop,
}

// sessionListCmd lists all active sessions
var sessionListCmd = &cobra.Command{
	Use:   "session-list",
	Short: "List all active browser sessions",
	Long: `List all active browser sessions.

OUTPUT:
  Returns a list of session IDs

EXAMPLES:
  browser-cli session-list`,
	RunE: runSessionList,
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
		var tabID int
		fmt.Sscanf(args[0], "%d", &tabID)
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
		tabID := 0
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &tabID)
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

	serverStartCmd.Flags().StringVar(&serverBrowser, "browser", "chromium", "Browser to use")
	serverStartCmd.Flags().BoolVar(&serverHeadless, "headless", true, "Run in headless mode")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Check if server already running
	client := browser.NewClient(serverSocket, sessionID)
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		fmt.Println("Server already running!")
		fmt.Printf("Session: %s\n", sessionID)
		fmt.Printf("Active tab: %v, Tabs: %v\n", resp.Data["active_tab"], resp.Data["tabs"])
		return nil
	}

	// Start new server
	fmt.Println("Starting browser server...")
	if sessionID != "" {
		fmt.Printf("Session: %s\n", sessionID)
	}

	cfg := browser.ServerConfig{
		Browser:    serverBrowser,
		Headless:   serverHeadless,
		SocketPath: serverSocket,
		SessionID:  sessionID,
		Proxy:      proxy,
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
		fmt.Println("\nShutting down server...")
		server.Stop()
		os.Exit(0)
	}()

	return server.Start()
}

func runServerStart(cmd *cobra.Command, args []string) error {
	// Check if server already running
	client := browser.NewClient(serverSocket, sessionID)
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		// Already running, just exit
		return nil
	}

	// Start new server
	cfg := browser.ServerConfig{
		Browser:    serverBrowser,
		Headless:   serverHeadless,
		SocketPath: serverSocket,
		SessionID:  sessionID,
		Proxy:      proxy,
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
	client := browser.NewClient("", sessionID)
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err != nil {
		return fmt.Errorf("server not running: %w", err)
	}

	output := map[string]interface{}{
		"command": "status",
		"status":  "success",
		"data":    resp.Data,
	}
	if sessionID != "" {
		output["session"] = sessionID
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
	return nil
}

func runStop(cmd *cobra.Command, args []string) error {
	client := browser.NewClient("", sessionID)
	resp, err := client.SendCommand(browser.Command{Action: "stop"})
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	if resp.Success {
		fmt.Println("Server stopped successfully")
		if sessionID != "" {
			fmt.Printf("Session: %s\n", sessionID)
		}
	}
	return nil
}

func runSessionList(cmd *cobra.Command, args []string) error {
	sessions, err := browser.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	output := map[string]interface{}{
		"command":  "session-list",
		"status":   "success",
		"sessions": sessions,
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
	return nil
}

// ensureServer ensures a browser server is running, starts one if not
func ensureServer() (*browser.Client, error) {
	client := browser.NewClient("", sessionID)

	// Check if server is already running
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		// Server already running
		return client, nil
	}

	// Server not running, start it in background
	serverCmd := exec.Command(os.Args[0], "server-start")
	if sessionID != "" {
		serverCmd.Args = append(serverCmd.Args, "--session", sessionID)
	}
	// Add headless flag based on current setting
	if headless {
		serverCmd.Args = append(serverCmd.Args, "--headless")
	} else {
		serverCmd.Args = append(serverCmd.Args, "--headless=false")
	}

	// Start server in background
	if err := serverCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to be ready (increased timeout for Playwright initialization)
	for i := 0; i < 60; i++ { // 60 * 500ms = 30 seconds
		time.Sleep(500 * time.Millisecond)
		resp, err = client.SendCommand(browser.Command{Action: "status"})
		if err == nil && resp.Success {
			return client, nil
		}
	}

	return nil, fmt.Errorf("server failed to start within 30 seconds")
}

// sendCommand sends a command to the server and outputs the result
func sendCommand(action string, params map[string]interface{}) error {
	client, err := ensureServer()
	if err != nil {
		return err
	}

	cmd := browser.Command{
		Action: action,
		Params: params,
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

	output := map[string]interface{}{
		"command": action,
		"status":  "success",
	}
	if resp.Data != nil {
		output["data"] = resp.Data
	}
	if resp.Error != "" {
		output["status"] = "error"
		output["error"] = resp.Error
	}
	if sessionID != "" {
		output["session"] = sessionID
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
	return nil
}