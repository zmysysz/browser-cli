package browser

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	Proxy      string
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
	var proxyServer *playwright.Proxy
	if cfg.Proxy != "" {
		proxyServer = &playwright.Proxy{Server: cfg.Proxy}
	}
	
	switch cfg.Browser {
	case "firefox":
		browser, err = pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
			Proxy:    proxyServer,
		})
	case "webkit":
		browser, err = pw.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
			Proxy:    proxyServer,
		})
	default:
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
			Proxy:    proxyServer,
			Args: []string{
				"--no-sandbox",
				"--disable-extensions",
				"--disable-default-apps",
			},
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

	// Read command with larger buffer
	buf := make([]byte, 262144) // 256KB buffer for large JSON requests
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

	return s.executeCommandInternal(cmd)
}

// executeCommandInternal executes without locking (called from executeCommand or run)
func (s *Server) executeCommandInternal(cmd Command) Response {
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
		err := page.Click(selector, playwright.PageClickOptions{
			Force: playwright.Bool(true),
		})
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "click-js":
		selector := cmd.Params["selector"].(string)
		page := s.tabs[s.activeTab]
		// Use JavaScript to click, bypassing Playwright's visibility checks
		script := fmt.Sprintf(`
			(function() {
				const el = document.querySelector('%s');
				if (el) {
					el.click();
					return 'clicked';
				}
				return 'not found';
			})();
		`, selector)
		result, err := page.Evaluate(script)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true, Data: map[string]interface{}{"result": result}}

	case "smart-click":
		selector := cmd.Params["selector"].(string)
		page := s.tabs[s.activeTab]
		// Use SmartClick to handle Web Components
		browser := &Browser{page: page}
		err := browser.SmartClick(selector, 30*time.Second)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "pick":
		x := cmd.Params["x"].(float64)
		y := cmd.Params["y"].(float64)
		depth := 5
		if d, ok := cmd.Params["depth"].(float64); ok {
			depth = int(d)
		}
		page := s.tabs[s.activeTab]
		
		pickScript := fmt.Sprintf(`
			(function() {
				const x = %f;
				const y = %f;
				const maxDepth = %d;
				
				const element = document.elementFromPoint(x, y);
				if (!element) return { success: false, error: 'No element found at coordinates' };
				
				// Helper: generate CSS selector for an element
				function generateSelector(el) {
					if (el.id) return '#' + el.id;
					if (el.tagName.includes('-')) return el.tagName.toLowerCase(); // Web Component
					let selector = el.tagName.toLowerCase();
					if (el.className && typeof el.className === 'string') {
						const classes = el.className.split(' ').filter(c => c && c.length < 30);
						if (classes.length > 0) selector += '.' + classes[0];
					}
					if (el.getAttribute('type')) selector += '[type="' + el.getAttribute('type') + '"]';
					if (el.getAttribute('name')) selector += '[name="' + el.getAttribute('name') + '"]';
					return selector;
				}
				
				// Helper: detect callable methods on element
				function detectMethods(obj) {
					const methods = [];
					const patterns = ['_on', '_handle', 'handle', 'on', '_click', '_submit', '_action'];
					try {
						for (const key of Object.keys(obj)) {
							if (typeof obj[key] === 'function') {
								for (const pattern of patterns) {
									if (key.toLowerCase().startsWith(pattern.toLowerCase())) {
										methods.push(key);
										break;
									}
								}
							}
						}
					} catch (e) {}
					return methods;
				}
				
				// Helper: get children summary (truncated)
				function getChildrenSummary(el, maxItems = 5) {
					const children = [];
					for (let i = 0; i < Math.min(el.children.length, maxItems); i++) {
						const child = el.children[i];
						children.push(generateSelector(child));
					}
					if (el.children.length > maxItems) {
						children.push('... (' + (el.children.length - maxItems) + ' more)');
					}
					return children;
				}
				
				// Helper: get attributes
				function getAttributes(el) {
					const attrs = {};
					for (const attr of el.attributes) {
						if (attr.value.length < 100) {
							attrs[attr.name] = attr.value;
						} else {
							attrs[attr.name] = attr.value.substring(0, 100) + '...';
						}
					}
					return attrs;
				}
				
				// Build target info
				const target = {
					tagName: element.tagName,
					selector: generateSelector(element),
					text: element.textContent ? element.textContent.substring(0, 50).trim() : '',
					attributes: getAttributes(element),
					methods: detectMethods(element),
					rect: {
						x: element.getBoundingClientRect().x,
						y: element.getBoundingClientRect().y,
						width: element.getBoundingClientRect().width,
						height: element.getBoundingClientRect().height
					}
				};
				
				// Build ancestor chain
				const ancestors = [];
				let current = element.parentElement;
				let level = 1;
				while (current && level <= maxDepth) {
					ancestors.push({
						level: level,
						tagName: current.tagName,
						selector: generateSelector(current),
						attributes: getAttributes(current),
						methods: detectMethods(current),
						children: getChildrenSummary(current)
					});
					current = current.parentElement;
					level++;
				}
				
				// Check for Shadow DOM
				let shadowDOM = null;
				if (element.shadowRoot) {
					shadowDOM = {
						host: generateSelector(element),
						children: getChildrenSummary(element.shadowRoot)
					};
				}
				
				// Generate suggestions
				const suggestions = [];
				if (element.tagName.includes('-')) {
					suggestions.push('Web Component detected: try smart-click or check methods');
				}
				if (target.methods.length > 0) {
					suggestions.push('Callable methods found: ' + target.methods.join(', '));
				}
				if (shadowDOM) {
					suggestions.push('Shadow DOM present: check shadowDOM.children for internal elements');
				}
				
				return {
					success: true,
					target: target,
					ancestors: ancestors,
					shadowDOM: shadowDOM,
					suggestions: suggestions
				};
			})();
		`, x, y, depth)
		
		result, err := page.Evaluate(pickScript)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		// Convert result to map
		if resultMap, ok := result.(map[string]interface{}); ok {
			return Response{Success: true, Data: resultMap}
		}
		return Response{Success: true, Data: map[string]interface{}{"result": result}}

	case "hover":
		selector := cmd.Params["selector"].(string)
		page := s.tabs[s.activeTab]
		
		// First, get the element's position
		posScript := fmt.Sprintf(`
			(function() {
				const el = document.querySelector('%s');
				if (!el) return null;
				const rect = el.getBoundingClientRect();
				return {x: rect.x + rect.width/2, y: rect.y + rect.height/2};
			})();
		`, selector)
		posResult, err := page.Evaluate(posScript)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		
		// Show virtual cursor at the position
		if posResult != nil {
			// Convert posResult to JSON string for embedding in script
			posJSON, _ := json.Marshal(posResult)
			cursorScript := fmt.Sprintf(`
				(function() {
					// Remove existing cursor
					const existing = document.getElementById('virtual-cursor');
					if (existing) existing.remove();
					
					// Create new cursor
					const cursor = document.createElement('div');
					cursor.id = 'virtual-cursor';
					cursor.style.cssText = 'position:fixed;width:20px;height:20px;background:red;border-radius:50%%;z-index:99999;pointer-events:none;transition:all 0.1s;';
					const pos = %s;
					cursor.style.left = pos.x + 'px';
					cursor.style.top = pos.y + 'px';
					document.body.appendChild(cursor);
					return 'cursor shown';
				})();
			`, string(posJSON))
			page.Evaluate(cursorScript)
		}
		
		// Perform hover
		err = page.Hover(selector, playwright.PageHoverOptions{
			Force: playwright.Bool(true),
		})
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
			Path:    playwright.String(path),
			Timeout: playwright.Float(5000), // 5 seconds timeout
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

	case "scroll":
		direction := cmd.Params["direction"].(string)
		distance := 300.0
		if d, ok := cmd.Params["distance"].(float64); ok {
			distance = d
		}
		page := s.tabs[s.activeTab]
		if direction == "up" {
			distance = -distance
		}
		page.Evaluate(fmt.Sprintf("window.scrollBy(0, %f)", distance))
		return Response{Success: true}

	case "back":
		page := s.tabs[s.activeTab]
		_, err := page.GoBack()
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "forward":
		page := s.tabs[s.activeTab]
		_, err := page.GoForward()
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "reload":
		page := s.tabs[s.activeTab]
		_, err := page.Reload()
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "select":
		selector := cmd.Params["selector"].(string)
		value := cmd.Params["value"].(string)
		page := s.tabs[s.activeTab]
		_, err := page.SelectOption(selector, playwright.SelectOptionValues{
			Values: &[]string{value},
		})
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		return Response{Success: true}

	case "elements":
		selector := cmd.Params["selector"].(string)
		page := s.tabs[s.activeTab]
		elements, err := page.QuerySelectorAll(selector)
		if err != nil {
			return Response{Success: false, Error: err.Error()}
		}
		var items []map[string]interface{}
		for _, el := range elements {
			tag, _ := el.Evaluate("el => el.tagName")
			text, _ := el.Evaluate("el => el.textContent")
			id, _ := el.Evaluate("el => el.id")
			class, _ := el.Evaluate("el => el.className")
			href, _ := el.Evaluate("el => el.href")
			visible, _ := el.IsVisible()
			items = append(items, map[string]interface{}{
				"tag":     tag,
				"text":    text,
				"id":      id,
				"class":   class,
				"href":    href,
				"visible": visible,
			})
		}
		return Response{Success: true, Data: map[string]interface{}{
			"count": len(items),
			"items": items,
		}}

	case "run":
		// Execute multiple actions in sequence
		actions, ok := cmd.Params["actions"].([]interface{})
		if !ok {
			return Response{Success: false, Error: "run requires actions array"}
		}
		results := make([]map[string]interface{}, 0)
		for i, actionInterface := range actions {
			actionMap, ok := actionInterface.(map[string]interface{})
			if !ok {
				results = append(results, map[string]interface{}{
					"step":   i + 1,
					"status": "error",
					"error":  "invalid action format",
				})
				continue
			}
			// Create sub-command
			subCmd := Command{
				Action: actionMap["action"].(string),
				Params: actionMap["params"].(map[string]interface{}),
			}
			// Execute without locking (already locked)
			resp := s.executeCommandInternal(subCmd)
			result := map[string]interface{}{
				"step":   i + 1,
				"action": subCmd.Action,
				"status": "success",
			}
			if resp.Success && resp.Data != nil {
				result["data"] = resp.Data
			}
			if !resp.Success {
				result["status"] = "error"
				result["error"] = resp.Error
			}
			results = append(results, result)
		}
		return Response{Success: true, Data: map[string]interface{}{
			"total_steps": len(results),
			"results":     results,
		}}

	case "stop":
		s.running = false
		// Stop will be called after response is sent
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.Stop()
		}()
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

	buf := make([]byte, 262144) // 256KB buffer for large JSON responses
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