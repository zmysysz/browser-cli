package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/browser-cli/internal/browser"
)

var (
	serverBrowser  string
	serverHeadless bool
	serverSocket   string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a persistent browser server",
	Long: `Start a persistent browser server that keeps the browser running.

The server listens for commands via Unix socket and maintains browser state
across multiple CLI invocations. This allows for continuous browser operations
without restarting the browser each time.

EXAMPLES:
  # Start server in headless mode (default)
  browser-cli server

  # Start server with visible browser window
  browser-cli server --headless=false

  # Start with specific browser
  browser-cli server --browser firefox

  # Check if server is running
  browser-cli status

  # Stop the server
  browser-cli stop

USAGE FOR AI:
  1. Start server first: browser-cli server --headless=false
  2. Execute commands: browser-cli exec navigate https://example.com
  3. Stop when done: browser-cli stop

SOCKET LOCATION:
  - $GAL_TMPDIR/browser-cli/server.sock (sandbox environment)
  - ~/.browser-cli/server.sock (normal environment)
`,
	RunE: runServer,
}

// statusCmd shows server status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check browser server status",
	Long: `Check if the browser server is running and show current state.

OUTPUT (JSON):
  {
    "running": true,
    "tabs": 2,
    "active_tab": 1
  }
`,
	RunE: runStatus,
}

// sessionListCmd lists all active sessions
var sessionListCmd = &cobra.Command{
	Use:   "session-list",
	Short: "List all active browser sessions",
	Long: `List all active browser sessions with their status.

OUTPUT (JSON):
  {
    "sessions": ["default", "task-1", "task-2"]
  }
`,
	RunE: runSessionList,
}

// stopCmd stops the server
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the browser server",
	Long: `Stop the browser server and close all browser instances.

This will:
  1. Save all cookies
  2. Close all tabs
  3. Close the browser
  4. Remove the socket file
`,
	RunE: runStop,
}

// execCmd sends a command to the server
var execCmd = &cobra.Command{
	Use:   "exec <action> [args...]",
	Short: "Execute a command on the browser server",
	Long: `Execute a command on the running browser server.

ACTIONS:
  navigate <url>          Navigate to URL
  click <selector>         Click element
  fill <selector> <value>  Fill input field
  type <selector> <text>   Type text into element
  eval <script>            Execute JavaScript
  screenshot [path]        Take screenshot
  text                     Get page text
  wait <selector>          Wait for element

TAB ACTIONS:
  tab-list                 List all tabs
  tab-new                   Create new tab
  tab-switch <id>           Switch to tab
  tab-close [id]            Close tab (current if no id)

DIALOG ACTIONS:
  dialog-status             Check if there's a pending dialog
  dialog-accept [value]     Accept the dialog (with value for prompt)
  dialog-dismiss            Dismiss the dialog

EXAMPLES:
  # Navigate
  browser-cli exec navigate https://example.com

  # Click element
  browser-cli exec click "button.submit"

  # Fill form
  browser-cli exec fill "#username" "user123"

  # Execute JavaScript
  browser-cli exec eval "document.title"

  # Take screenshot
  browser-cli exec screenshot /tmp/page.png

  # Tab management
  browser-cli exec tab-new
  browser-cli exec tab-switch 2
  browser-cli exec tab-list
  browser-cli exec tab-close
`,
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(sessionListCmd)

	// Server flags
	serverCmd.Flags().StringVar(&serverBrowser, "browser", "chromium", "Browser to use (chromium, firefox, webkit)")
	serverCmd.Flags().BoolVar(&serverHeadless, "headless", true, "Run in headless mode")
	serverCmd.Flags().StringVar(&serverSocket, "socket", "", "Unix socket path (default: ~/.browser-cli/server.sock)")
}

func runServer(cmd *cobra.Command, args []string) error {
	// First check if server is already running
	client := browser.NewClient(serverSocket, sessionID)
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err == nil && resp.Success {
		// Server already running, show status
		fmt.Println("Server already running!")
		fmt.Printf("Session: %s\n", sessionID)
		fmt.Printf("Active tab: %v, Tabs: %v\n", resp.Data["active_tab"], resp.Data["tabs"])
		fmt.Println("Use 'browser-cli exec' to send commands")
		fmt.Println("Use 'browser-cli stop' to stop the server")
		return nil
	}

	// Server not running, start new one
	fmt.Println("Starting new browser server...")
	if sessionID != "" {
		fmt.Printf("Session: %s\n", sessionID)
	}
	cfg := browser.ServerConfig{
		Browser:    serverBrowser,
		Headless:   serverHeadless,
		SocketPath: serverSocket,
		SessionID:  sessionID,
	}

	server, err := browser.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
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

func runStatus(cmd *cobra.Command, args []string) error {
	client := browser.NewClient("", sessionID)
	resp, err := client.SendCommand(browser.Command{Action: "status"})
	if err != nil {
		return fmt.Errorf("Server not running: %w", err)
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
		return fmt.Errorf("Failed to stop server: %w", err)
	}

	if resp.Success {
		fmt.Println("Server stopped successfully")
		if sessionID != "" {
			fmt.Printf("Session: %s\n", sessionID)
		}
	}
	return nil
}

func runExec(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Usage: browser-cli exec <action> [args...]")
	}

	action := args[0]
	client := browser.NewClient("", sessionID)

	var command browser.Command

	switch action {
	case "navigate":
		if len(args) < 2 {
			return fmt.Errorf("Usage: browser-cli exec navigate <url>")
		}
		command = browser.Command{
			Action: "navigate",
			Params: map[string]interface{}{"url": args[1]},
		}

	case "click":
		if len(args) < 2 {
			return fmt.Errorf("Usage: browser-cli exec click <selector>")
		}
		command = browser.Command{
			Action: "click",
			Params: map[string]interface{}{"selector": args[1]},
		}

	case "fill":
		if len(args) < 3 {
			return fmt.Errorf("Usage: browser-cli exec fill <selector> <value>")
		}
		command = browser.Command{
			Action: "fill",
			Params: map[string]interface{}{
				"selector": args[1],
				"value":    args[2],
			},
		}

	case "type":
		if len(args) < 3 {
			return fmt.Errorf("Usage: browser-cli exec type <selector> <text>")
		}
		command = browser.Command{
			Action: "type",
			Params: map[string]interface{}{
				"selector": args[1],
				"text":     args[2],
			},
		}

	case "eval":
		if len(args) < 2 {
			return fmt.Errorf("Usage: browser-cli exec eval <script>")
		}
		command = browser.Command{
			Action: "eval",
			Params: map[string]interface{}{"script": args[1]},
		}

	case "screenshot":
		path := "screenshot.png"
		if len(args) >= 2 {
			path = args[1]
		}
		command = browser.Command{
			Action: "screenshot",
			Params: map[string]interface{}{"path": path},
		}

	case "text":
		command = browser.Command{Action: "text"}

	case "wait":
		if len(args) < 2 {
			return fmt.Errorf("Usage: browser-cli exec wait <selector>")
		}
		command = browser.Command{
			Action: "wait",
			Params: map[string]interface{}{"selector": args[1]},
		}

	// Tab actions
	case "tab-list":
		command = browser.Command{Action: "tab_list"}

	case "tab-new":
		command = browser.Command{Action: "tab_new"}

	case "tab-switch":
		if len(args) < 2 {
			return fmt.Errorf("Usage: browser-cli exec tab-switch <id>")
		}
		var tabID int
		fmt.Sscanf(args[1], "%d", &tabID)
		command = browser.Command{
			Action: "tab_switch",
			TabID:  tabID,
		}

	case "tab-close":
		tabID := 0
		if len(args) >= 2 {
			fmt.Sscanf(args[1], "%d", &tabID)
		}
		command = browser.Command{
			Action: "tab_close",
			TabID:  tabID,
		}

	// Dialog actions
	case "dialog-status":
		command = browser.Command{Action: "dialog_status"}

	case "dialog-accept":
		value := ""
		if len(args) >= 2 {
			value = args[1]
		}
		command = browser.Command{
			Action: "dialog_accept",
			Params: map[string]interface{}{"value": value},
		}

	case "dialog-dismiss":
		command = browser.Command{Action: "dialog_dismiss"}

	default:
		return fmt.Errorf("Unknown action: %s", action)
	}

	resp, err := client.SendCommand(command)
	if err != nil {
		return fmt.Errorf("Command failed: %w", err)
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

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
	return nil
}

func runSessionList(cmd *cobra.Command, args []string) error {
	sessions, err := browser.ListSessions()
	if err != nil {
		return fmt.Errorf("Failed to list sessions: %w", err)
	}

	output := map[string]interface{}{
		"command": "session-list",
		"status":  "success",
		"sessions": sessions,
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
	return nil
}