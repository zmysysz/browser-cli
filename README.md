# Browser-CLI

[English](README.md) | [中文](README_zh.md)

A powerful command-line tool for browser automation, designed specifically for AI agents and automated workflows.

## Why Browser-CLI?

- **AI-First Design** - Structured JSON output, clear command semantics, perfect for LLM integration
- **Auto-Managed Server** - Browser server starts automatically when needed, no manual management
- **Multi-Agent Support** - Isolated sessions for parallel agent execution
- **Cookie Persistence** - Automatic cookie save/load, maintain login states
- **Dialog Detection** - Detect and handle JavaScript dialogs with AI decision-making
- **Cross-Browser** - Chromium, Firefox, WebKit support via Playwright

## Installation

```bash
# Clone and build
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build

# Install Playwright browsers (first time only)
make setup-browsers

# Install to system
make install
```

## Quick Start

### Simple Commands (Auto Server Management)

The browser server starts automatically when you run any command:

```bash
# Navigate - server auto-starts
browser-cli navigate https://example.com

# Take screenshot - uses existing server
browser-cli screenshot /tmp/page.png

# Extract text
browser-cli text

# Stop server when done
browser-cli stop
```

### Run Command (Multi-step Operations)

Execute multiple actions in a single command:

```bash
browser-cli run "navigate https://example.com; click a; text"
```

## All Commands

### Global Options

```
--browser, -b     Browser: chromium, firefox, webkit (default: chromium)
--headless        Headless mode (default: true)
--timeout, -t     Timeout duration (default: 30s)
--output, -o      Output format: json, markdown (default: markdown)
--session, -s     Session ID for isolated browser instance (required)
--proxy           Proxy server URL (e.g. http://proxy.example.com:8080 or socks5://proxy:1080)
```

### Navigation Commands

| Command | Description |
|---------|-------------|
| `navigate <url>` | Navigate to URL |
| `back` | Go back in history |
| `forward` | Go forward in history |
| `reload` | Reload current page |

### Interaction Commands

| Command | Description |
|---------|-------------|
| `click <selector>` | Click element |
| `click-js <selector>` | Click using JavaScript (bypasses visibility checks) |
| `smart-click <selector>` | Intelligently click Web Components (auto-detects internal methods) |
| `hover <selector>` | Hover over element (shows virtual cursor) |
| `fill <selector> <value>` | Fill input field |
| `type <selector> <text>` | Type text character by character |
| `select <selector> <value>` | Select dropdown option |
| `eval <script>` | Execute JavaScript |
| `pick <x> <y> [--depth=N]` | Pick element at coordinates, return DOM hierarchy and methods |
| `right-click <selector>` | Right-click element (context menu) |
| `dblclick <selector>` | Double-click element |
| `upload <selector> <file>` | Upload file to file input |
| `keyboard <key>` | Press keyboard key/combo (e.g. Ctrl+A, Enter) |

### Web Component Support

Browser-CLI provides special commands for handling Web Components (custom elements):

#### `smart-click` - Auto-detect and click Web Components

Web Components often use internal callback functions instead of standard DOM events. `smart-click` automatically detects and calls these methods:

```bash
# Works with custom elements like <custom-button>, <xhs-publish-btn>
browser-cli smart-click "custom-button"
browser-cli smart-click "[data-action=publish]"
```

Detection patterns: `_on*`, `_handle*`, `handle*`, `_click`, `_submit`, `_action`

#### `pick` - Discover element internals

Use `pick` to inspect elements at specific coordinates and discover their internal structure:

```bash
# Pick element at coordinates (from screenshot)
browser-cli pick 500 300 --depth=5

# Returns:
{
  "target": {
    "tagName": "CUSTOM-BUTTON",
    "selector": "custom-button",
    "methods": ["_onClick", "_onPublish"],  // Detected callable methods!
    "attributes": {"data-action": "publish"}
  },
  "ancestors": [
    {"level": 1, "tagName": "DIV", "selector": ".toolbar", "children": ["custom-button", "save-btn"]}
  ],
  "shadowDOM": {"host": "custom-button", "children": ["button.internal"]},
  "suggestions": ["Web Component detected: try smart-click"]
}
```

Use cases:
- Discover internal methods like `_onPublish` on Web Components
- Find correct selector for nested elements
- Understand Shadow DOM structure
- Debug why `click()` doesn't work on custom elements

## File Upload

Upload files to file input elements:

```bash
# Upload a single file
browser-cli upload "#file-input" ./document.pdf

# Upload to a specific input type
browser-cli upload "input[type=file]" /tmp/image.png
```

## PDF Export

Save the current page as a PDF file (Chromium only):

```bash
# Save as PDF with default settings (A4, portrait)
browser-cli pdf

# Custom output path
browser-cli pdf report.pdf

# Landscape orientation and Letter format
browser-cli pdf --landscape --format Letter page.pdf
```

## Keyboard Shortcuts

Press keyboard keys and key combinations:

```bash
# Single keys
browser-cli keyboard "Enter"
browser-cli keyboard "Escape"
browser-cli keyboard "Tab"

# Key combinations
browser-cli keyboard "Ctrl+A"        # Select all
browser-cli keyboard "Ctrl+C"        # Copy
browser-cli keyboard "Ctrl+V"        # Paste
browser-cli keyboard "Ctrl+Shift+I"  # DevTools
```

### Supported Keys

| Category | Keys |
|----------|------|
| Single | Enter, Escape, Tab, Backspace, Delete, Space |
| Modifiers | Ctrl+, Alt+, Shift+, Meta+ |
| Arrows | ArrowUp, ArrowDown, ArrowLeft, ArrowRight |
| Function | F1-F12 |

## Advanced Click

Beyond the basic `click` command, Browser-CLI supports right-click and double-click:

```bash
# Right-click to open context menu
browser-cli right-click "#menu-item"
browser-cli right-click ".context-target"

# Double-click to select or activate
browser-cli dblclick "#item"
browser-cli dblclick "td.editable"
```

### Extraction Commands

| Command | Description |
|---------|-------------|
| `screenshot [path]` | Take screenshot |
| `text` | Extract page text |
| `elements <selector>` | Find elements |
| `pdf [file]` | Save page as PDF (Chromium only, flags: --landscape, --format) |

### Utility Commands

| Command | Description |
|---------|-------------|
| `wait <selector>` | Wait for element |
| `scroll <direction>` | Scroll page (up/down) |

### Tab Commands

| Command | Description |
|---------|-------------|
| `tab-new` | Create new tab |
| `tab-switch <id>` | Switch to tab |
| `tab-list` | List all tabs |
| `tab-close [id]` | Close tab |

### Dialog Commands

| Command | Description |
|---------|-------------|
| `dialog-status` | Check pending dialog |
| `dialog-accept [value]` | Accept dialog |
| `dialog-dismiss` | Dismiss dialog |

### Server Commands

| Command | Description |
|---------|-------------|
| `server` | Start server manually (foreground) |
| `status` | Check server status |
| `stop` | Stop server and save cookies |
| `session-list` | List all active sessions |

### Cookie Commands

| Command | Description |
|---------|-------------|
| `cookie list` | List saved cookies |
| `cookie clear [domain]` | Clear cookies |
| `cookie clear --all` | Clear all cookies |

## Multi-Session Support

Run multiple isolated browser sessions for parallel agent execution:

```bash
# Each session has independent browser instance
browser-cli --session agent-1 navigate https://example.com
browser-cli --session agent-2 navigate https://google.com

# List all sessions
browser-cli session-list

# Stop specific session
browser-cli --session agent-1 stop
```

### Session Isolation

Each session has:
- Independent browser instance
- Independent tab management
- Shared cookie storage (login states preserved)

## Cookie Management

Cookies are automatically saved and loaded, maintaining login states across sessions:

```bash
# View saved cookies
browser-cli cookie list

# Clear cookies for a domain
browser-cli cookie clear example.com

# Clear all cookies
browser-cli cookie clear --all
```

### Cookie Storage

Cookies are stored in `/tmp/browser-cli/cookies/<domain>.json`

## Dialog Detection

Browser-CLI detects JavaScript dialogs and lets AI decide how to handle them:

```bash
# Check for pending dialog
browser-cli dialog-status
# Returns: {"dialog": {"type": "confirm", "message": "Are you sure?"}}

# Accept dialog
browser-cli dialog-accept

# Dismiss dialog
browser-cli dialog-dismiss

# Accept prompt with value
browser-cli dialog-accept "user input"
```

### Supported Dialog Types

| Type | Description |
|------|-------------|
| `alert` | Simple alert, can only accept |
| `confirm` | Yes/No dialog, accept or dismiss |
| `prompt` | Input dialog, accept with value |
| `beforeunload` | Page leave confirmation |

Note: Custom HTML popups (like privacy policy dialogs) are not detected. Use element selectors to handle them.

## Output Formats

### JSON (Recommended for AI)

```json
{
  "command": "navigate",
  "status": "success",
  "data": {
    "url": "https://example.com/",
    "title": "Example Domain"
  }
}
```

### Markdown (Human-readable)

```
## Navigate
- Status: success
- URL: https://example.com/
- Title: Example Domain
```

## Selector Syntax

Browser-CLI supports multiple selector formats:

| Selector | Example | Description |
|----------|---------|-------------|
| CSS | `#username` | CSS selector |
| Text | `text=Submit` | Element containing text |
| Role | `role=button` | Element by ARIA role |
| XPath | `xpath=//div[@id="main"]` | XPath selector |

## AI Integration Best Practices

1. **Auto Server** - Server starts automatically, just run commands
2. **Use JSON Output** - Parse results programmatically with `--output json`
3. **Check Status Field** - "success" or "error" for each operation
4. **Handle Dialogs** - Check `dialog-status` before proceeding
5. **Use Sessions** - Isolate parallel agent tasks with `--session`
6. **Preserve Login** - Cookies auto-save, login states persist

### Example AI Workflow

```bash
# 1. Navigate and login (server auto-starts)
browser-cli navigate https://login.example.com
browser-cli fill "#username" "user"
browser-cli fill "#password" "pass"
browser-cli click "button[type=submit]"

# 2. Wait for login
browser-cli wait ".dashboard"

# 3. Press keyboard shortcut
browser-cli keyboard "Enter"

# 4. Check for dialogs
browser-cli dialog-status

# 5. Extract data
browser-cli text
browser-cli screenshot result.png

# 6. Stop server (cookies auto-saved)
browser-cli stop
```

## Storage Paths

| Resource | Path |
|----------|------|
| Cookies | `/tmp/browser-cli/cookies/<domain>.json` |
| Sessions | `/tmp/browser-cli/sessions/<session-id>/server.sock` |

## Requirements

- Go 1.21+
- Playwright browsers (installed via `make setup-browsers`)

## License

[Apache License 2.0](LICENSE)

Copyright 2024 zmysysz