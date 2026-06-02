# Browser-CLI Skill

Browser automation via command-line interface. Use this skill when the user asks to browse websites, interact with web pages, take screenshots, fill forms, extract web content, or automate any browser task.

## Prerequisites

Browser-CLI must be installed and available in PATH:

```bash
# Install from source
git clone https://github.com/zmysysz/browser-cli
cd browser-cli && make build && make install

# Install Playwright browsers (first time only)
make setup-browsers
```

Verify installation:
```bash
browser-cli status
```

## Core Concepts

- **Auto-managed server**: The browser server starts automatically on first command. No manual start needed.
- **Session isolation**: Use `--session <id>` to run independent browser instances for parallel tasks.
- **JSON output**: Use `--output json` for structured results (default for AI agents).
- **Proxy support**: Use `--proxy http://host:port` if behind a proxy.

## Command Reference

### Navigation
```bash
browser-cli navigate <url>           # Go to URL
browser-cli back                     # Go back
browser-cli forward                  # Go forward
browser-cli reload                   # Reload page
```

### Interaction
```bash
browser-cli click <selector>         # Click element
browser-cli click-js <selector>      # JS click (bypasses visibility)
browser-cli smart-click <selector>   # Click Web Components
browser-cli right-click <selector>   # Right-click (context menu)
browser-cli dblclick <selector>      # Double-click
browser-cli hover <selector>         # Hover over element
browser-cli fill <selector> <value>  # Fill input field
browser-cli type <selector> <text>   # Type text character by character
browser-cli select <selector> <val>  # Select dropdown option
browser-cli keyboard <key>           # Press key/combo (Ctrl+A, Enter, etc.)
browser-cli upload <selector> <file> # Upload file
```

### Extraction
```bash
browser-cli text                     # Extract page text
browser-cli screenshot [path]        # Take screenshot
browser-cli elements <selector>      # Find elements
browser-cli pdf [file]               # Save as PDF (Chromium only)
browser-cli eval <javascript>        # Execute JavaScript
```

### Utility
```bash
browser-cli wait <selector>          # Wait for element
browser-cli scroll <up|down>         # Scroll page
browser-cli pick <x> <y> [--depth=N] # Inspect element at coordinates
```

### Tabs
```bash
browser-cli tab-new                  # New tab
browser-cli tab-list                 # List tabs
browser-cli tab-switch <id>          # Switch tab
browser-cli tab-close [id]           # Close tab
```

### Dialogs
```bash
browser-cli dialog-status            # Check for pending dialog
browser-cli dialog-accept [value]    # Accept dialog
browser-cli dialog-dismiss           # Dismiss dialog
```

### Cookies & Sessions
```bash
browser-cli cookie list              # List saved cookies
browser-cli cookie clear [domain]    # Clear cookies
browser-cli --session <id> <cmd>     # Run in isolated session
browser-cli session-list             # List active sessions
browser-cli stop                     # Stop server (cookies auto-saved)
```

### Multi-step
```bash
browser-cli run "navigate <url>; click <sel>; text"
```

## Selector Syntax

| Type | Example | Description |
|------|---------|-------------|
| CSS | `#username` | Standard CSS selector |
| Text | `text=Submit` | Element containing text |
| Role | `role=button` | ARIA role selector |
| XPath | `xpath=//div[@id]` | XPath expression |

## Common Patterns

### Login to a website
```bash
browser-cli navigate https://example.com/login
browser-cli fill "#username" "user@example.com"
browser-cli fill "#password" "secret"
browser-cli click "button[type=submit]"
browser-cli wait ".dashboard"
```

### Extract data from a page
```bash
browser-cli navigate https://example.com/data
browser-cli text
browser-cli eval "JSON.stringify(Array.from(document.querySelectorAll('table tr')).map(tr => tr.textContent))"
```

### Screenshot and PDF
```bash
browser-cli navigate https://example.com/report
browser-cli screenshot report.png
browser-cli pdf --landscape report.pdf
```

### Handle a dialog
```bash
browser-cli click "button.delete"
browser-cli dialog-status
# If dialog detected:
browser-cli dialog-accept
```

### Multi-tab workflow
```bash
browser-cli navigate https://example.com
browser-cli tab-new
browser-cli tab-switch 2
browser-cli navigate https://other.com
browser-cli tab-switch 1
```

### Use with proxy
```bash
browser-cli --proxy http://10.10.42.134:7890 navigate https://example.com
```

## Output Format

All commands return JSON with `status` field:
```json
{"command": "navigate", "status": "success", "data": {"url": "...", "title": "..."}}
{"command": "click", "status": "error", "error": "element not found"}
```

Always check `status` === "success" before proceeding.

## Important Notes

- Server auto-starts on first command — no need to manually start it
- Always call `browser-cli stop` when done to clean up resources
- Use `--headless=false` for debugging (shows browser window)
- Use `--session <id>` to isolate parallel tasks
- Cookies persist across sessions automatically
- `pdf` command only works with Chromium browser
- `smart-click` is for Web Components that don't respond to normal clicks
- `pick` command helps discover element selectors from coordinates
