# AGENTS.md — Browser-CLI

This project provides `browser-cli`, a command-line browser automation tool for AI agents.

## Tool: browser-cli

A Go CLI that wraps Playwright to let AI agents control browsers via shell commands.

### Installation

```bash
git clone https://github.com/zmysysz/browser-cli
cd browser-cli && make build && make install
make setup-browsers  # first time only
```

### How It Works

- Client-server architecture over Unix sockets
- Server auto-starts on first command, no manual management needed
- Each `--session <id>` gets an isolated browser instance
- Cookies persist automatically across sessions
- All output is JSON with `status: success|error`

### Command Quick Reference

```bash
# Navigation
browser-cli navigate <url>
browser-cli back | forward | reload

# Interaction
browser-cli click <selector>
browser-cli fill <selector> <value>
browser-cli type <selector> <text>
browser-cli select <selector> <value>
browser-cli hover <selector>
browser-cli right-click <selector>
browser-cli dblclick <selector>
browser-cli keyboard <key>              # "Ctrl+A", "Enter", "Tab"
browser-cli upload <selector> <file>
browser-cli smart-click <selector>      # Web Components

# Extraction
browser-cli text                        # page text
browser-cli screenshot [path]
browser-cli elements <selector>
browser-cli eval <javascript>
browser-cli pdf [file]                  # Chromium only

# Utility
browser-cli wait <selector>
browser-cli scroll up|down
browser-cli pick <x> <y> [--depth=N]   # inspect element at coords

# Tabs
browser-cli tab-new | tab-list | tab-switch <id> | tab-close [id]

# Dialogs
browser-cli dialog-status | dialog-accept [value] | dialog-dismiss

# Sessions
browser-cli --session <id> <cmd>        # isolated session
browser-cli stop                        # cleanup
```

### Selectors

| Type | Syntax | Example |
|------|--------|---------|
| CSS | standard | `#login-btn`, `input[name=email]` |
| Text | `text=...` | `text=Submit` |
| Role | `role=...` | `role=button` |
| XPath | `xpath=...` | `xpath=//div[@id="main"]` |

### Important Flags

- `--output json` — structured output (recommended for agents)
- `--session <id>` — isolated browser context for parallel work
- `--proxy http://host:port` — proxy for network access
- `--headless=false` — show browser window for debugging
- `--timeout 60s` — operation timeout

### Typical Workflow

```bash
browser-cli navigate https://example.com/login
browser-cli fill "#email" "user@example.com"
browser-cli fill "#password" "secret"
browser-cli click "button[type=submit]"
browser-cli wait ".dashboard"
browser-cli text
browser-cli screenshot result.png
browser-cli stop
```

### Notes

- Always call `browser-cli stop` when done to free resources
- Use `--session` for parallel agent tasks
- `smart-click` handles Web Components that ignore normal clicks
- `pick` helps discover element selectors from screenshot coordinates
- `pdf` only works with Chromium browser
