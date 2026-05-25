# Browser-CLI

A powerful command-line tool for browser automation, designed specifically for AI agents and automated workflows.

## Why Browser-CLI?

- **AI-First Design** - Structured JSON output, clear command semantics, perfect for LLM integration
- **Persistent Browser Sessions** - Keep browser alive across multiple commands with server mode
- **Multi-Agent Support** - Isolated sessions for parallel agent execution
- **Cookie Persistence** - Automatic cookie save/load, maintain login states
- **Dialog Detection** - Detect and handle JavaScript dialogs with AI decision-making
- **Cross-Browser** - Chromium, Firefox, WebKit support via Playwright

## Installation

```bash
# Clone and build
git clone https://github.com/browser-cli/browser-cli
cd browser-cli
make build

# Install Playwright browsers (first time only)
make setup-browsers

# Install to system
make install
```

## Quick Start

### Server Mode (Recommended)

Start a persistent browser server for continuous operations:

```bash
# Start server with visible browser
browser-cli server --headless=false

# Execute commands
browser-cli exec navigate https://example.com
browser-cli exec screenshot /tmp/page.png
browser-cli exec text

# Stop server when done
browser-cli stop
```

### Single Commands

For one-off operations without server:

```bash
browser-cli navigate https://example.com
browser-cli screenshot page.png
browser-cli text
```

### Run Command (Multi-step Operations)

Execute multiple actions in a single browser session:

```bash
browser-cli run "navigate https://example.com; click a; text"
```

## Server Mode

Server mode keeps the browser running, allowing multiple commands to execute in the same browser instance.

### Commands

```bash
# Start server
browser-cli server --headless=false

# Check status
browser-cli status

# Execute actions
browser-cli exec navigate https://example.com
browser-cli exec click "button.submit"
browser-cli exec fill "#username" "user123"
browser-cli exec screenshot /tmp/page.png
browser-cli exec text

# Tab management
browser-cli exec tab-new
browser-cli exec tab-switch 2
browser-cli exec tab-list
browser-cli exec tab-close

# Dialog handling
browser-cli exec dialog-status
browser-cli exec dialog-accept
browser-cli exec dialog-dismiss

# Stop server
browser-cli stop
```

### Server Options

| Option | Description |
|--------|-------------|
| `--headless` | Run without visible window (default: true) |
| `--browser` | Browser engine: chromium, firefox, webkit |

## Multi-Session Support

Run multiple isolated browser sessions for parallel agent execution:

```bash
# Start multiple sessions
browser-cli --session agent-1 server --headless=false
browser-cli --session agent-2 server --headless=false

# List all sessions
browser-cli session-list

# Execute in specific session
browser-cli --session agent-1 exec navigate https://example.com
browser-cli --session agent-2 exec navigate https://google.com

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
browser-cli exec dialog-status
# Returns: {"dialog": {"type": "confirm", "message": "Are you sure?"}}

# Accept dialog
browser-cli exec dialog-accept

# Dismiss dialog
browser-cli exec dialog-dismiss

# Accept prompt with value
browser-cli exec dialog-accept "user input"
```

### Supported Dialog Types

| Type | Description |
|------|-------------|
| `alert` | Simple alert, can only accept |
| `confirm` | Yes/No dialog, accept or dismiss |
| `prompt` | Input dialog, accept with value |
| `beforeunload` | Page leave confirmation |

Note: Custom HTML popups (like privacy policy dialogs) are not detected. Use element selectors to handle them.

## All Commands

### Global Options

```
--browser, -b     Browser: chromium, firefox, webkit (default: chromium)
--headless        Headless mode (default: true)
--timeout, -t     Timeout duration (default: 30s)
--output, -o      Output format: json, markdown (default: markdown)
--session, -s     Session ID for isolated browser instance
```

### Server Commands

| Command | Description |
|---------|-------------|
| `server` | Start persistent browser server |
| `status` | Check server status |
| `stop` | Stop server and save cookies |
| `session-list` | List all active sessions |

### Exec Actions

| Action | Description |
|--------|-------------|
| `navigate <url>` | Navigate to URL |
| `click <selector>` | Click element |
| `fill <selector> <value>` | Fill input field |
| `type <selector> <text>` | Type text |
| `eval <script>` | Execute JavaScript |
| `screenshot [path]` | Take screenshot |
| `text` | Extract page text |
| `elements <selector>` | Find elements |
| `wait <selector>` | Wait for element |
| `scroll <direction>` | Scroll page (up/down) |

### Tab Actions

| Action | Description |
|--------|-------------|
| `tab-new` | Create new tab |
| `tab-switch <id>` | Switch to tab |
| `tab-list` | List all tabs |
| `tab-close [id]` | Close tab |

### Dialog Actions

| Action | Description |
|--------|-------------|
| `dialog-status` | Check pending dialog |
| `dialog-accept [value]` | Accept dialog |
| `dialog-dismiss` | Dismiss dialog |

### Cookie Commands

| Command | Description |
|---------|-------------|
| `cookie list` | List saved cookies |
| `cookie clear [domain]` | Clear cookies |
| `cookie clear --all` | Clear all cookies |

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

1. **Use Server Mode** - Start server, execute multiple commands, stop when done
2. **Use JSON Output** - Parse results programmatically
3. **Check Status Field** - "success" or "error" for each operation
4. **Handle Dialogs** - Check `dialog-status` before proceeding
5. **Use Sessions** - Isolate parallel agent tasks with `--session`
6. **Preserve Login** - Cookies auto-save, login states persist

### Example AI Workflow

```bash
# 1. Start server
browser-cli server --headless=false

# 2. Navigate and login
browser-cli exec navigate https://login.example.com
browser-cli exec fill "#username" "user"
browser-cli exec fill "#password" "pass"
browser-cli exec click "button[type=submit]"

# 3. Wait for login
browser-cli exec wait ".dashboard"

# 4. Check for dialogs
browser-cli exec dialog-status

# 5. Extract data
browser-cli exec text
browser-cli exec screenshot result.png

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

MIT