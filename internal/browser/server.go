package browser

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
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

// SessionState holds per-session browser state (one BrowserContext per session)
type SessionState struct {
	ID            string
	Context       playwright.BrowserContext
	Tabs          map[int]playwright.Page
	ActiveTab     int
	NextTabID     int
	PendingDialog *DialogInfo
	DialogChan    chan playwright.Dialog
}

// Server manages a single browser instance with multiple session contexts
type Server struct {
	mu           sync.Mutex
	pw           *playwright.Playwright
	browser      playwright.Browser
	sessions     map[string]*SessionState // sessionID -> SessionState
	listener     net.Listener
	socketPath   string
	running      bool
	lastActivity time.Time
	idleTimeout  time.Duration
	stopOnce     sync.Once

	// connWG tracks in-flight handleConnection goroutines so Stop can
	// wait for them to finish writing their response before tearing
	// the listener down. This replaces the old time.Sleep(100ms) hack.
	connWG sync.WaitGroup

	// idleDisabled is true when IdleTimeout <= 0 was passed in. The
	// idle monitor checks this flag instead of using a fake 1h default.
	idleDisabled bool

	// Browser config (shared across all sessions)
	browserType string
	headless    bool
	proxy       string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Browser     string
	Headless    bool
	SocketPath  string
	Proxy       string
	IdleTimeout time.Duration
}

// Command represents a command sent to the server
type Command struct {
	Action    string                 `json:"action"`
	SessionID string                 `json:"session_id,omitempty"`
	TabID     int                    `json:"tab_id,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
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

	// Determine socket path — single socket for all sessions
	socketPath := cfg.SocketPath
	if socketPath == "" {
		socketPath = GetSocketPath("")
	}
	os.MkdirAll(filepath.Dir(socketPath), 0755)

	// Remove existing socket file
	os.Remove(socketPath)

	// Create listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	idleTimeout := cfg.IdleTimeout
	idleDisabled := false
	if idleTimeout <= 0 {
		idleDisabled = true
		// Keep a non-zero value so the math in idleMonitor stays well-defined
		// in case some path inspects it; the flag short-circuits the check.
		idleTimeout = 1 * time.Hour
	}

	server := &Server{
		pw:           pw,
		browser:      browser,
		sessions:     make(map[string]*SessionState),
		listener:     listener,
		socketPath:   socketPath,
		running:      true,
		lastActivity: time.Now(),
		idleTimeout:  idleTimeout,
		idleDisabled: idleDisabled,
		browserType:  cfg.Browser,
		headless:     cfg.Headless,
		proxy:        cfg.Proxy,
	}

	// Start idle timeout monitor
	go server.idleMonitor()

	return server, nil
}

// idleMonitor checks for idle timeout and shuts down the server if no commands
// have been received within the configured idle period.
func (s *Server) idleMonitor() {
	if s.idleDisabled {
		return
	}
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		if !s.running {
			s.mu.Unlock()
			return
		}
		idle := time.Since(s.lastActivity)
		s.mu.Unlock()

		if idle >= s.idleTimeout {
			fmt.Printf("Server idle for %v (threshold: %v), shutting down...\n", idle.Round(time.Minute), s.idleTimeout)
			s.Stop()
			return
		}
	}
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

		s.connWG.Add(1)
		go func(c net.Conn) {
			defer s.connWG.Done()
			s.handleConnection(c)
		}(conn)
	}

	// After the Accept loop has unwound, no new connections will be
	// accepted. Stop the server (idempotent via stopOnce).
	s.Stop()
	return nil
}

// maxFrameBytes caps any single request/response body. Anything larger
// is almost certainly a misuse (e.g. sending a base64'd PDF inline).
const maxFrameBytes = 16 * 1024 * 1024 // 16 MiB

// readFrame reads exactly one length-prefixed message from conn.
//
// Wire format: 4-byte big-endian length, then `length` raw bytes.
//
// We avoid bufio.Scanner's default 64KiB limit and the old single-read
// 256KiB buffer (which silently truncated large payloads). Reads are
// driven by io.ReadFull so short reads do not corrupt the message.
func readFrame(conn net.Conn) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[:])
	if length == 0 {
		return nil, fmt.Errorf("empty frame")
	}
	if length > maxFrameBytes {
		return nil, fmt.Errorf("frame too large: %d bytes (max %d)", length, maxFrameBytes)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// writeFrame writes a single length-prefixed message to conn.
func writeFrame(conn net.Conn, data []byte) error {
	if len(data) > maxFrameBytes {
		return fmt.Errorf("frame too large: %d bytes (max %d)", len(data), maxFrameBytes)
	}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	if _, err := conn.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	body, err := readFrame(conn)
	if err != nil {
		// Connection closed before any frame arrived, or a bad frame.
		// We have nothing to write back to, so just return.
		return
	}

	var cmd Command
	if err := json.Unmarshal(body, &cmd); err != nil {
		s.sendResponse(conn, Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid command: %v", err),
		})
		return
	}

	// Default session ID
	if cmd.SessionID == "" {
		cmd.SessionID = "default"
	}

	// Touch lastActivity so the idle monitor doesn't shut us down while
	// we're actively processing requests.
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()

	// Execute command
	resp := s.executeCommand(cmd)
	s.sendResponse(conn, resp)
}

// fail builds a failure Response with err's message. Centralising the
// format here means future additions (request id, error code, stack
// trace) only need to touch one place.
func fail(err error) Response {
	return Response{Success: false, Error: err.Error()}
}

// getSession returns an existing session or creates a new one
func (s *Server) getSession(sessionID string) (*SessionState, error) {
	if ss, ok := s.sessions[sessionID]; ok {
		return ss, nil
	}

	// Create new BrowserContext
	context, err := s.browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context for session %s: %w", sessionID, err)
	}

	// Auto-load saved cookies for this session
	cookieStorage := NewSessionCookieStorage(sessionID)
	cookies, err := cookieStorage.LoadAll()
	if err == nil && len(cookies) > 0 {
		optionalCookies := make([]playwright.OptionalCookie, len(cookies))
		for i, c := range cookies {
			optionalCookies[i] = c.ToOptionalCookie()
		}
		if err := context.AddCookies(optionalCookies); err != nil {
			fmt.Printf("Warning: failed to load cookies for session %s: %v\n", sessionID, err)
		} else {
			fmt.Printf("Loaded %d cookies for session %s\n", len(cookies), sessionID)
		}
	}

	// Create initial page
	page, err := context.NewPage()
	if err != nil {
		context.Close()
		return nil, fmt.Errorf("failed to create page for session %s: %w", sessionID, err)
	}

	ss := &SessionState{
		ID:         sessionID,
		Context:    context,
		Tabs:       map[int]playwright.Page{1: page},
		ActiveTab:  1,
		NextTabID:  2,
		DialogChan: make(chan playwright.Dialog, 10),
	}

	// Setup dialog handler for initial tab
	s.setupDialogHandler(ss, page)

	s.sessions[sessionID] = ss
	return ss, nil
}

// setupDialogHandler sets up dialog detection for a page within a session
func (s *Server) setupDialogHandler(ss *SessionState, page playwright.Page) {
	page.On("dialog", func(dialog playwright.Dialog) {
		s.mu.Lock()
		ss.PendingDialog = &DialogInfo{
			Type:         dialog.Type(),
			Message:      dialog.Message(),
			DefaultValue: dialog.DefaultValue(),
			TabID:        ss.ActiveTab,
		}
		ss.DialogChan <- dialog
		s.mu.Unlock()
	})
}

// executeCommand executes a command and returns response
func (s *Server) executeCommand(cmd Command) Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update last activity time
	s.lastActivity = time.Now()

	return s.executeCommandInternal(cmd)
}

// executeCommandInternal executes without locking (called from executeCommand or run)
func (s *Server) executeCommandInternal(cmd Command) Response {
	switch cmd.Action {
	case "ping":
		return Response{Success: true, Data: map[string]interface{}{"status": "ok"}}

	case "status":
		sessionList := make([]string, 0, len(s.sessions))
		for id := range s.sessions {
			sessionList = append(sessionList, id)
		}
		return Response{
			Success: true,
			Data: map[string]interface{}{
				"running":       s.running,
				"sessions":      sessionList,
				"session_count": len(s.sessions),
				"idle_timeout":  s.idleTimeout.String(),
				"last_activity": s.lastActivity.Format(time.RFC3339),
			},
		}

	// Session management commands
	case "session_list":
		sessionList := make([]string, 0, len(s.sessions))
		for id := range s.sessions {
			sessionList = append(sessionList, id)
		}
		return Response{Success: true, Data: map[string]interface{}{
			"sessions": sessionList,
			"count":    len(sessionList),
		}}

	case "session_status":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		return Response{
			Success: true,
			Data: map[string]interface{}{
				"session":     cmd.SessionID,
				"tabs":        len(ss.Tabs),
				"active_tab":  ss.ActiveTab,
				"has_dialog":  ss.PendingDialog != nil,
			},
		}

	case "session_close":
		ss, ok := s.sessions[cmd.SessionID]
		if !ok {
			return Response{Success: false, Error: fmt.Sprintf("Session %s not found", cmd.SessionID)}
		}
		// Save cookies before closing
		s.saveSessionCookies(ss)
		// Close all tabs and context
		for _, page := range ss.Tabs {
			page.Close()
		}
		ss.Context.Close()
		delete(s.sessions, cmd.SessionID)
		return Response{Success: true, Data: map[string]interface{}{
			"message": fmt.Sprintf("Session %s closed", cmd.SessionID),
		}}

	// Tab commands — require a session
	case "tab_list":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		var tabList []map[string]interface{}
		for id, page := range ss.Tabs {
			tabList = append(tabList, map[string]interface{}{
				"id":    id,
				"url":   page.URL(),
				"title": func() string { t, _ := page.Title(); return t }(),
			})
		}
		return Response{Success: true, Data: map[string]interface{}{"tabs": tabList}}

	case "tab_new":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		page, err := ss.Context.NewPage()
		if err != nil {
			return fail(err)
		}
		s.setupDialogHandler(ss, page)
		id := ss.NextTabID
		ss.Tabs[id] = page
		ss.NextTabID++
		ss.ActiveTab = id
		return Response{Success: true, Data: map[string]interface{}{"tab_id": id}}

	case "tab_switch":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		tabID := cmd.TabID
		if _, ok := ss.Tabs[tabID]; !ok {
			return Response{Success: false, Error: fmt.Sprintf("Tab %d not found", tabID)}
		}
		ss.ActiveTab = tabID
		return Response{Success: true, Data: map[string]interface{}{"active_tab": tabID}}

	case "tab_close":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		tabID := cmd.TabID
		if tabID == 0 {
			tabID = ss.ActiveTab
		}
		if tabID == ss.ActiveTab && len(ss.Tabs) == 1 {
			return Response{Success: false, Error: "Cannot close the last active tab"}
		}
		if page, ok := ss.Tabs[tabID]; ok {
			page.Close()
			delete(ss.Tabs, tabID)
			if tabID == ss.ActiveTab {
				// Pick the next active tab deterministically: lowest id
				// among the remaining tabs. Map iteration order is
				// randomized in Go, which would make this command
				// non-reproducible across runs.
				ids := make([]int, 0, len(ss.Tabs))
				for id := range ss.Tabs {
					ids = append(ids, id)
				}
				sort.Ints(ids)
				if len(ids) > 0 {
					ss.ActiveTab = ids[0]
				}
			}
		}
		return Response{Success: true}

	// Navigation commands
	case "navigate":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		url, err := paramString(cmd.Params, "url")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.Goto(url)
		if err != nil {
			return fail(err)
		}
		title, _ := page.Title()
		return Response{
			Success: true,
			Data: map[string]interface{}{
				"url":   page.URL(),
				"title": title,
			},
		}

	case "back":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.GoBack()
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "forward":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.GoForward()
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "reload":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.Reload()
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	// Interaction commands
	case "click":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		err = page.Click(selector, playwright.PageClickOptions{
			Force: playwright.Bool(true),
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "click-js":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
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
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"result": result}}

	case "smart-click":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		browser := &Browser{page: page}
		err = browser.SmartClick(selector, 30*time.Second)
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "pick":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		x, err := paramFloat(cmd.Params, "x")
		if err != nil {
			return fail(err)
		}
		y, err := paramFloat(cmd.Params, "y")
		if err != nil {
			return fail(err)
		}
		depth := int(optFloat(cmd.Params, "depth", 5))
		page := ss.Tabs[ss.ActiveTab]

		pickScript := fmt.Sprintf(`
			(function() {
				const x = %f;
				const y = %f;
				const maxDepth = %d;
				
				const element = document.elementFromPoint(x, y);
				if (!element) return { success: false, error: 'No element found at coordinates' };
				
				function generateSelector(el) {
					if (el.id) return '#' + el.id;
					if (el.tagName.includes('-')) return el.tagName.toLowerCase();
					let selector = el.tagName.toLowerCase();
					if (el.className && typeof el.className === 'string') {
						const classes = el.className.split(' ').filter(c => c && c.length < 30);
						if (classes.length > 0) selector += '.' + classes[0];
					}
					if (el.getAttribute('type')) selector += '[type="' + el.getAttribute('type') + '"]';
					if (el.getAttribute('name')) selector += '[name="' + el.getAttribute('name') + '"]';
					return selector;
				}
				
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
				
				let shadowDOM = null;
				if (element.shadowRoot) {
					shadowDOM = {
						host: generateSelector(element),
						children: getChildrenSummary(element.shadowRoot)
					};
				}
				
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
			return fail(err)
		}
		if resultMap, ok := result.(map[string]interface{}); ok {
			return Response{Success: true, Data: resultMap}
		}
		return Response{Success: true, Data: map[string]interface{}{"result": result}}

	case "hover":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]

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
			return fail(err)
		}

		if posResult != nil {
			posJSON, _ := json.Marshal(posResult)
			cursorScript := fmt.Sprintf(`
				(function() {
					const existing = document.getElementById('virtual-cursor');
					if (existing) existing.remove();
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

		err = page.Hover(selector, playwright.PageHoverOptions{
			Force: playwright.Bool(true),
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "fill":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		value, err := paramString(cmd.Params, "value")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		err = page.Fill(selector, value)
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "type":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		text, err := paramString(cmd.Params, "text")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		err = page.Type(selector, text)
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "select":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		value, err := paramString(cmd.Params, "value")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.SelectOption(selector, playwright.SelectOptionValues{
			Values: &[]string{value},
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	// Extraction commands
	case "eval":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		script, err := paramString(cmd.Params, "script")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		result, err := page.Evaluate(script)
		if err != nil {
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"value": result}}

	case "screenshot":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		path := optString(cmd.Params, "path", "screenshot.png")
		if path == "" {
			path = "screenshot.png"
		}
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.Screenshot(playwright.PageScreenshotOptions{
			Path:    playwright.String(path),
			Timeout: playwright.Float(5000),
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"path": path}}

	case "text":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		text, err := page.InnerText("body")
		if err != nil {
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"text": text}}

	case "elements":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		elements, err := page.QuerySelectorAll(selector)
		if err != nil {
			return fail(err)
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

	// Wait/scroll
	case "wait":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		timeout := optFloat(cmd.Params, "timeout", 30000.0)
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(timeout),
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "scroll":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		direction, err := paramString(cmd.Params, "direction")
		if err != nil {
			return fail(err)
		}
		distance := optFloat(cmd.Params, "distance", 300.0)
		page := ss.Tabs[ss.ActiveTab]
		if direction == "up" {
			distance = -distance
		}
		page.Evaluate(fmt.Sprintf("window.scrollBy(0, %f)", distance))
		return Response{Success: true}

	case "upload":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		path, err := paramString(cmd.Params, "path")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		el, err := page.QuerySelector(selector)
		if err != nil {
			return fail(err)
		}
		if el == nil {
			return Response{Success: false, Error: "Element not found: " + selector}
		}
		err = el.SetInputFiles(path)
		if err != nil {
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"path": path}}

	case "pdf":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		path := optString(cmd.Params, "path", "output.pdf")
		landscape := optBool(cmd.Params, "landscape", false)
		format := optString(cmd.Params, "format", "A4")
		page := ss.Tabs[ss.ActiveTab]
		_, err = page.PDF(playwright.PagePdfOptions{
			Path:      playwright.String(path),
			Landscape: playwright.Bool(landscape),
			Format:    playwright.String(format),
		})
		if err != nil {
			err = fmt.Errorf("PDF generation failed (note: only supported in Chromium): %w", err)
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"path": path}}

	case "keyboard":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		key, err := paramString(cmd.Params, "key")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		err = page.Keyboard().Press(key)
		if err != nil {
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{"key": key}}

	case "right-click":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		err = page.Click(selector, playwright.PageClickOptions{
			Button: playwright.MouseButtonRight,
			Force:  playwright.Bool(true),
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	case "dblclick":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		selector, err := paramString(cmd.Params, "selector")
		if err != nil {
			return fail(err)
		}
		page := ss.Tabs[ss.ActiveTab]
		err = page.Dblclick(selector, playwright.PageDblclickOptions{
			Force: playwright.Bool(true),
		})
		if err != nil {
			return fail(err)
		}
		return Response{Success: true}

	// Dialog commands
	case "dialog_status":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		if ss.PendingDialog == nil {
			return Response{Success: true, Data: map[string]interface{}{"dialog": nil}}
		}
		return Response{Success: true, Data: map[string]interface{}{"dialog": ss.PendingDialog}}

	case "dialog_accept":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		if ss.PendingDialog == nil {
			return Response{Success: false, Error: "No pending dialog"}
		}
		select {
		case dialog := <-ss.DialogChan:
			value := optString(cmd.Params, "value", "")
			if value != "" && ss.PendingDialog.Type == "prompt" {
				dialog.Accept(value)
			} else {
				dialog.Accept()
			}
			ss.PendingDialog = nil
			return Response{Success: true, Data: map[string]interface{}{"action": "accepted"}}
		default:
			return Response{Success: false, Error: "Dialog event not received"}
		}

	case "dialog_dismiss":
		ss, err := s.getSession(cmd.SessionID)
		if err != nil {
			return fail(err)
		}
		if ss.PendingDialog == nil {
			return Response{Success: false, Error: "No pending dialog"}
		}
		select {
		case dialog := <-ss.DialogChan:
			dialog.Dismiss()
			ss.PendingDialog = nil
			return Response{Success: true, Data: map[string]interface{}{"action": "dismissed"}}
		default:
			return Response{Success: false, Error: "Dialog event not received"}
		}

	// Cookie commands
	case "cookie_list":
		cookieStorage := NewSessionCookieStorage(cmd.SessionID)
		infos, err := cookieStorage.List()
		if err != nil {
			return fail(err)
		}
		return Response{Success: true, Data: map[string]interface{}{
			"domains": infos,
			"count":   len(infos),
		}}

	case "cookie_clear":
		cookieStorage := NewSessionCookieStorage(cmd.SessionID)
		domain := optString(cmd.Params, "domain", "")
		clearAll := optBool(cmd.Params, "all", false)
		if domain == "" && !clearAll {
			return Response{Success: false, Error: "specify a domain or use all=true"}
		}
		if err := cookieStorage.Clear(domain); err != nil {
			return fail(err)
		}
		msg := fmt.Sprintf("Cookies cleared for session %s", cmd.SessionID)
		if clearAll {
			msg = fmt.Sprintf("All cookies cleared for session %s", cmd.SessionID)
		} else {
			msg = fmt.Sprintf("Cookies cleared for %s in session %s", domain, cmd.SessionID)
		}
		return Response{Success: true, Data: map[string]interface{}{"message": msg}}

	// Run multi-step
	case "run":
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
			actionName, err := paramString(actionMap, "action")
			if err != nil {
				results = append(results, map[string]interface{}{
					"step":   i + 1,
					"status": "error",
					"error":  "invalid action format: " + err.Error(),
				})
				continue
			}
			params, _ := actionMap["params"].(map[string]interface{})

			// Handle sleep locally: it's a server-side pause between actions,
			// not a real browser command. This avoids racing with the next
			// request that the client would otherwise send in parallel.
			if actionName == "sleep" {
				var durationMs float64
				if params != nil {
					if v, ok := params["duration_ms"]; ok {
						switch n := v.(type) {
						case float64:
							durationMs = n
						case int:
							durationMs = float64(n)
						default:
							results = append(results, map[string]interface{}{
								"step":   i + 1,
								"action": actionName,
								"status": "error",
								"error":  fmt.Sprintf("sleep: duration_ms must be number, got %T", v),
							})
							continue
						}
					}
				}
				if durationMs > 0 {
					time.Sleep(time.Duration(durationMs) * time.Millisecond)
				}
				results = append(results, map[string]interface{}{
					"step":   i + 1,
					"action": actionName,
					"status": "success",
					"data":   map[string]interface{}{"duration_ms": durationMs},
				})
				continue
			}

			subCmd := Command{
				Action:    actionName,
				SessionID: cmd.SessionID,
				Params:    params,
			}
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
		// Mark the server as stopping, close the listener to unblock
		// Accept(), and let Start() drive the actual teardown via Stop()
		// once the loop has unwound. We deliberately do not call Stop()
		// from inside handleConnection — that would race with Stop()
		// being called by idleMonitor or a signal handler, and it would
		// run before this response is written to the client.
		s.running = false
		if s.listener != nil {
			s.listener.Close()
		}
		return Response{Success: true, Data: map[string]interface{}{"message": "Server stopped"}}

	default:
		return Response{Success: false, Error: fmt.Sprintf("Unknown action: %s", cmd.Action)}
	}
}

// saveSessionCookies saves cookies for a session before it's closed
func (s *Server) saveSessionCookies(ss *SessionState) {
	if ss.Context == nil {
		return
	}
	cookies, err := ss.Context.Cookies()
	if err == nil && len(cookies) > 0 {
		cookieStorage := NewSessionCookieStorage(ss.ID)
		cookieStorage.SaveAll(cookies)
	}
}

// sendResponse sends a response to the client, framed with a length prefix
// so large responses (e.g. page text, element lists) are not truncated.
func (s *Server) sendResponse(conn net.Conn, resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		// Best-effort fallback: still frame it.
		_ = writeFrame(conn, []byte(`{"success":false,"error":"Failed to marshal response"}`))
		return
	}
	if err := writeFrame(conn, data); err != nil {
		fmt.Fprintf(os.Stderr, "sendResponse: write failed: %v\n", err)
	}
}

// Stop stops the server and cleans up all resources.
// Safe to call multiple times — only executes once via stopOnce.
//
// When called from a request handler (e.g. the "stop" action), the handler
// is still in-flight in its own goroutine; we wait for it (and any other
// in-flight handlers) to finish writing their response before tearing the
// listener and browser down.
func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		s.running = false

		// Close listener (unblocks Accept() in Start()).
		if s.listener != nil {
			s.listener.Close()
		}

		// Wait for all in-flight handlers to finish so their responses
		// are flushed before we tear sockets/contexts down.
		s.connWG.Wait()

		// Save cookies and close all sessions
		for id, ss := range s.sessions {
			s.saveSessionCookies(ss)
			for _, page := range ss.Tabs {
				page.Close()
			}
			if ss.Context != nil {
				ss.Context.Close()
			}
			delete(s.sessions, id)
		}

		// Close browser
		if s.browser != nil {
			s.browser.Close()
		}

		// Stop Playwright
		if s.pw != nil {
			s.pw.Stop()
		}

		// Remove socket file
		os.Remove(s.socketPath)
	})
}

// GetSocketPath returns the socket path for the server
// In single-server mode, there's only one socket regardless of session
func GetSocketPath(sessionID string) string {
	return filepath.Join("/tmp", "browser-cli", "server.sock")
}

// Client connects to the browser server
type Client struct {
	socketPath string
}

// NewClient creates a new client
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		socketPath = GetSocketPath("")
	}
	return &Client{socketPath: socketPath}
}

// SendCommand sends a command to the server and waits for the response.
//
// Wire format on the unix socket: 4-byte big-endian length prefix followed
// by the JSON payload. The framing is the same on the way back. This lets
// requests and responses be arbitrarily large (up to 16 MiB) without
// truncation.
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

	if err := writeFrame(conn, data); err != nil {
		return Response{}, fmt.Errorf("Failed to send command: %w", err)
	}

	body, err := readFrame(conn)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return Response{}, fmt.Errorf("Failed to unmarshal response: %w", err)
	}

	return resp, nil
}

// Ping checks if server is running
func (c *Client) Ping() bool {
	resp, err := c.SendCommand(Command{Action: "ping"})
	return err == nil && resp.Success
}
