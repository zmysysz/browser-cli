package browser

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// DialogInfo represents a pending dialog
type DialogInfo struct {
	Type         string `json:"type"`          // alert, confirm, prompt, beforeunload
	Message      string `json:"message"`       // Dialog message text
	DefaultValue string `json:"default_value"` // Default value for prompt
	TabID        int    `json:"tab_id"`        // Which tab has the dialog
}

// Server manages a persistent browser instance
type Server struct {
	mu           sync.Mutex
	pw           *playwright.Playwright
	browser      playwright.Browser
	context      playwright.BrowserContext
	tabs         map[int]playwright.Page
	activeTab    int
	nextTabID    int
	listener     net.Listener
	socketPath   string
	running      bool
	pendingDialog *DialogInfo          // Current pending dialog
	dialogChan   chan playwright.Dialog // Channel for dialog events
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Browser    string
	Headless   bool
	SocketPath string
	MaxTabs    int
	SessionID  string
}

// Command represents a command sent to the server
type Command struct {
	Action   string                 `json:"action"`
	TabID    int                    `json:"tab_id,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// Response represents a response from the server
type Response struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// NewServer creates a new browser server
func NewServer(cfg ServerConfig) (*Server, error) {
	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to init playwright: %w", err)
	}

	// Launch browser
	var browser playwright.Browser
	switch cfg.Browser {
	case "firefox":
		browser, err = pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
		})
	case "webkit":
		browser, err = pw.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
		})
	default:
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
		})
	}
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Create context with dialog handler
	context, err := browser.NewContext()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	// Auto-load saved cookies
	cookies, err := GetCookieStorage().LoadAll()
	if err == nil && len(cookies) > 0 {
		optionalCookies := make([]playwright.OptionalCookie, len(cookies))
		for i, c := range cookies {
			optionalCookies[i] = c.ToOptionalCookie()
		}
		if err := context.AddCookies(optionalCookies); err != nil {
			fmt.Printf("Warning: failed to load cookies: %v\n", err)
		} else {
			fmt.Printf("Loaded %d cookies from storage\n", len(cookies))
		}
	}

	// Create initial tab
	page, err := context.NewPage()
	if err != nil {
		context.Close()
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Determine socket path
	socketPath := cfg.SocketPath
	if socketPath == "" {
		socketPath = GetSocketPath(cfg.SessionID)
	}
	os.MkdirAll(filepath.Dir(socketPath), 0755)

	// Remove existing socket file
	os.Remove(socketPath)

	// Create listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		page.Close()
		context.Close()
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	server := &Server{
		pw:           pw,
		browser:      browser,
		context:      context,
		tabs:         map[int]playwright.Page{1: page},
		activeTab:    1,
		nextTabID:    2,
		listener:     listener,
		socketPath:   socketPath,
		running:      true,
		dialogChan:   make(chan playwright.Dialog, 10),
	}

	// Setup dialog handler for initial tab (detect and notify, don't auto-handle)
	page.On("dialog", func(dialog playwright.Dialog) {
		server.mu.Lock()
		server.pendingDialog = &DialogInfo{
			Type:         dialog.Type(),
			Message:      dialog.Message(),
			DefaultValue: dialog.DefaultValue(),
			TabID:        server.activeTab,
		}
		// Send dialog to channel for later handling
		server.dialogChan <- dialog
		server.mu.Unlock()
	})

	return server, nil
}

// Start starts the server and listens for commands
func (s *Server) Start() error {
	fmt.Printf("Browser server started at %s\n", s.socketPath)
	fmt.Println("Press Ctrl+C to stop")

	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				fmt.Printf("Accept error: %v\n", err)
			}
			continue
		}

		go s.handleConnection(conn)
	}

	return nil
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read command
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var cmd Command
	if err := json.Unmarshal(buf[:n], &cmd); err != nil {
		s.sendResponse(conn, Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid command: %v", err),
		})
		return
	}

	// Execute command
	resp := s.executeCommand(cmd)
	s.sendResponse(conn, resp)
}

// executeCommand executes a command and returns response
func (s *Server) executeCommand(cmd Command) Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch cmd.Action {
	case "ping":
		return Response{Success: true, Data: map[string]interface{}{"status": "ok"}}

	case "status":
		return Response{
			Success: true,
			Data: map[string]interface{}{
				"running":    s.running,
				"tabs":       len(s.tabs),
				"active_tab": s.activeTab,
			},
		}

	case "tab_list":
		var tabList []map[string]interface{}
		for id, page := range s.tabs {
			tabList = append(tabList, map[string]interface{}{
				"id":    id,
				"url":   page.URL(),
				"title": func() string { t, _ := page.Title(); return t }(),
			})
		}
		return Response{Success: true, Data: map[string]interface{}{"tabs": tabList}}

	case "tab_new":
		page, err := s.context.NewPage()
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		// Setup dialog handler (detect and notify)
		page.On("dialog", func(dialog playwright.Dialog) {
			s.mu.Lock()
			s.pendingDialog = &DialogInfo{
				Type:         dialog.Type(),
				Message:      dialog.Message(),
				DefaultValue: dialog.DefaultValue(),
				TabID:        s.activeTab,
			}
			s.dialogChan <- dialog
			s.mu.Unlock()
		})
		id := s.nextTabID
		s.tabs[id] = page
		s.nextTabID++
		s.activeTab = id
		return Response{Success: true, Data: map[string]interface{}{"tab_id": id}}

	case "tab_switch":
		tabID := cmd.TabID
		if _, ok := s.tabs[tabID]; !ok {
			return Response{Success: false, Error: fmt.Sprintf("Tab %d not found", tabID)}
		}
		s.activeTab = tabID
		return Response{Success: true, Data: map[string]interface{}{"active_tab": tabID}}

	case "tab_close":
		tabID := cmd.TabID
		if tabID == 0 {
			tabID = s.activeTab
		}
		if tabID == s.activeTab && len(s.tabs) == 1 {
			return Response{Success: false, Error: "Cannot close the last active tab"}
		}
		if page, ok := s.tabs[tabID]; ok {
			page.Close()
			delete(s.tabs, tabID)
			if tabID == s.activeTab {
				// Switch to another tab
				for id := range s.tabs {
					s.activeTab = id
					break
				}
			}
		}
		return Response{Success: true}

	case "navigate":
		url := cmd.Params["url"].(string)
		page := s.tabs[s.activeTab]
		_, err := page.Goto(url)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		title, _ := page.Title()
		return Response{
			Success: true,
			Data: map[string]interface{}{
				"url":   page.URL(),
				"title": title,
			},
		}

	case "click":
		selector := cmd.Params["selector"].(string)
		page := s.tabs[s.activeTab]
		err := page.Click(selector)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "fill":
		selector := cmd.Params["selector"].(string)
		value := cmd.Params["value"].(string)
		page := s.tabs[s.activeTab]
		err := page.Fill(selector, value)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "type":
		selector := cmd.Params["selector"].(string)
		text := cmd.Params["text"].(string)
		page := s.tabs[s.activeTab]
		err := page.Type(selector, text)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "eval":
		script := cmd.Params["script"].(string)
		page := s.tabs[s.activeTab]
		result, err := page.Evaluate(script)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true, Data: map[string]interface{}{"value": result}}

	case "screenshot":
		path := cmd.Params["path"].(string)
		if path == "" {
			path = "screenshot.png"
		}
		page := s.tabs[s.activeTab]
		_, err := page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String(path),
		})
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true, Data: map[string]interface{}{"path": path}}

	case "text":
		page := s.tabs[s.activeTab]
		text, err := page.InnerText("body")
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true, Data: map[string]interface{}{"text": text}}

	case "wait":
		selector := cmd.Params["selector"].(string)
		timeout := 30000.0
		if t, ok := cmd.Params["timeout"].(float64); ok {
			timeout = t
		}
		page := s.tabs[s.activeTab]
		_, err := page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(timeout),
		})
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	// Dialog commands
	case "dialog_status":
		if s.pendingDialog == nil {
			return Response{Success: true, Data: map[string]interface{}{"dialog": nil}}
		}
		return Response{Success: true, Data: map[string]interface{}{"dialog": s.pendingDialog}}

	case "dialog_accept":
		if s.pendingDialog == nil {
			return Response{Success: false, Error: "No pending dialog"}
		}
		// Get dialog from channel
		select {
		case dialog := <-s.dialogChan:
			value := ""
			if v, ok := cmd.Params["value"].(string); ok {
				value = v
			}
			if value != "" && s.pendingDialog.Type == "prompt" {
				dialog.Accept(value)
			} else {
				dialog.Accept()
			}
			s.pendingDialog = nil
			return Response{Success: true, Data: map[string]interface{}{"action": "accepted"}}
		default:
			return Response{Success: false, Error: "Dialog event not received"}
		}

	case "dialog_dismiss":
		if s.pendingDialog == nil {
			return Response{Success: false, Error: "No pending dialog"}
		}
		// Get dialog from channel
		select {
		case dialog := <-s.dialogChan:
			dialog.Dismiss()
			s.pendingDialog = nil
			return Response{Success: true, Data: map[string]interface{}{"action": "dismissed"}}
		default:
			return Response{Success: false, Error: "Dialog event not received"}
		}

	case "stop":
		s.running = false
		s.Stop()
		return Response{Success: true, Data: map[string]interface{}{"message": "Server stopped"}}

	default:
		return Response{Success: false, Error: fmt.Sprintf("Unknown action: %s", cmd.Action)}
	}
}

// sendResponse sends a response to the client
func (s *Server) sendResponse(conn net.Conn, resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		conn.Write([]byte(`{"success":false,"error":"Failed to marshal response"}`))
		return
	}
	conn.Write(data)
}

// Stop stops the server
func (s *Server) Stop() {
	s.running = false

	// Save cookies
	if s.context != nil {
		cookies, err := s.context.Cookies()
		if err == nil && len(cookies) > 0 {
			GetCookieStorage().SaveAll(cookies)
		}
	}

	// Close all tabs
	for _, page := range s.tabs {
		page.Close()
	}

	// Close browser
	if s.browser != nil {
		s.browser.Close()
	}

	// Stop Playwright
	if s.pw != nil {
		s.pw.Stop()
	}

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Remove socket file
	os.Remove(s.socketPath)
}

// GetSocketPath returns the socket path for a session
func GetSocketPath(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join("/tmp", "browser-cli", "sessions", sessionID, "server.sock")
}

// Client connects to the browser server
type Client struct {
	socketPath string
}

// NewClient creates a new client
func NewClient(socketPath string, sessionID string) *Client {
	if socketPath == "" {
		socketPath = GetSocketPath(sessionID)
	}
	return &Client{socketPath: socketPath}
}

// SendCommand sends a command to the server
func (c *Client) SendCommand(cmd Command) (Response, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to connect to server: %w (is server running?)", err)
	}
	defer conn.Close()

	data, err := json.Marshal(cmd)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to marshal command: %w", err)
	}

	conn.Write(data)

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		return Response{}, fmt.Errorf("Failed to unmarshal response: %w", err)
	}

	return resp, nil
}

// Ping checks if server is running
func (c *Client) Ping() bool {
	resp, err := c.SendCommand(Command{Action: "ping"})
	return err == nil && resp.Success
}