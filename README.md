# Browser-CLI

[English](README.md) | [中文](README_zh.md)

A command-line tool for browser automation, designed for AI agents. Built with Go + Playwright.

**One command to control the browser — perfect for Codex, Claude Code, Cursor, and any AI coding assistant.**

```bash
browser-cli navigate https://example.com
browser-cli fill "#search" "browser automation"
browser-cli click "button[type=submit]"
browser-cli text
```

## Features

- 🤖 **AI-First** — Structured JSON output, clear command semantics, auto-managed server
- 🔒 **Session Isolation** — Each agent gets its own browser instance via `--session`
- 🍪 **Cookie Persistence** — Auto save/load, login states preserved across sessions
- 🌐 **Proxy Support** — `--proxy http://host:port` for network-restricted environments
- 🎯 **Web Components** — `smart-click` and `pick` for custom elements and Shadow DOM
- ⌨️ **Full Keyboard** — Shortcuts, combos, Tab/Enter/Escape, Ctrl+A/C/V
- 📄 **PDF & Screenshot** — Export pages as PDF or PNG
- 📁 **File Upload** — Upload files to any `<input type="file">`

## Quick Install

```bash
# Clone and build (requires Go 1.21+)
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build

# Install Playwright browsers (first time only)
make setup-browsers

# Add to PATH
make install

# Or build without CGO (fully static binary)
make build-static
```

> **No CGO required.** Browser-CLI compiles with `CGO_ENABLED=0` and supports cross-compilation:
> ```bash
> CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/browser-cli.exe .
> CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/browser-cli-mac .
> ```

## 30-Second Setup for AI Agents

### Claude Code

Copy the integration file to your project:

```bash
# Project-level (recommended)
mkdir -p .claude/commands/
cp integrations/claude/browser.md .claude/commands/

# Or user-level (all projects)
mkdir -p ~/.claude/commands/
cp integrations/claude/browser.md ~/.claude/commands/
```

Then in Claude Code, just ask: *"Navigate to github.com and take a screenshot"*

### OpenAI Codex

Copy the skill file:

```bash
# User-level
mkdir -p ~/.codex/skills/
cp integrations/codex/browser-cli.md ~/.codex/skills/
```

### Cursor / Windsurf / Any Agent

`AGENTS.md` is at the project root — most AI coding tools read it automatically.

### GAL (Global Agent Layer)

```bash
cp -r skills/browser-cli/ ~/.gal/skills/
```

> 📁 Integration templates are in `integrations/` — copy them to the tool's config directory.

## Quick Start

The browser server starts automatically — just run commands:

```bash
# Navigate (server auto-starts)
browser-cli navigate https://example.com

# Interact
browser-cli fill "#username" "user"
browser-cli click "button[type=submit]"

# Extract
browser-cli text
browser-cli screenshot page.png

# Done
browser-cli stop
```

### Multi-step in One Command

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
--session, -s     Session ID for isolated browser instance
--proxy           Proxy server URL (e.g. http://proxy:8080 or socks5://proxy:1080)
--idle-timeout    Auto-shutdown after idle period (default: 1h, 0 to disable)
```

### Navigation

| Command | Description |
|---------|-------------|
| `navigate <url>` | Navigate to URL |
| `back` | Go back in history |
| `forward` | Go forward in history |
| `reload` | Reload current page |

### Interaction

| Command | Description |
|---------|-------------|
| `click <selector>` | Click element |
| `click-js <selector>` | Click using JavaScript (bypasses visibility checks) |
| `smart-click <selector>` | Click Web Components (auto-detects internal methods) |
| `hover <selector>` | Hover over element |
| `fill <selector> <value>` | Fill input field |
| `type <selector> <text>` | Type text character by character |
| `select <selector> <value>` | Select dropdown option |
| `right-click <selector>` | Right-click (context menu) |
| `dblclick <selector>` | Double-click element |
| `keyboard <key>` | Press key/combo (e.g. `Ctrl+A`, `Enter`, `Tab`) |
| `upload <selector> <file>` | Upload file to file input |
| `eval <script>` | Execute JavaScript |
| `eval-file <path>` | Execute JavaScript read from a file (multi-line / quoted strings friendly) |

### Extraction

| Command | Description |
|---------|-------------|
| `text` | Extract page text |
| `screenshot [path]` | Take screenshot |
| `elements <selector>` | Find elements |
| `pdf [file]` | Save as PDF (Chromium only, flags: `--landscape`, `--format`) |

### Utility

| Command | Description |
|---------|-------------|
| `wait <selector>` | Wait for element to appear |
| `scroll <up\|down>` | Scroll page |
| `pick <x> <y> [--depth=N]` | Inspect element at coordinates, return DOM hierarchy |

### Tabs

| Command | Description |
|---------|-------------|
| `tab-new` | Create new tab |
| `tab-list` | List all tabs |
| `tab-switch <id>` | Switch to tab |
| `tab-close [id]` | Close tab |

### Dialogs

| Command | Description |
|---------|-------------|
| `dialog-status` | Check pending dialog |
| `dialog-accept [value]` | Accept dialog |
| `dialog-dismiss` | Dismiss dialog |

### Cookies & Sessions

| Command | Description |
|---------|-------------|
| `cookie list` | List saved cookies |
| `cookie clear [domain]` | Clear cookies for domain |
| `cookie clear --all` | Clear all cookies |
| `session-list` | List active sessions |
| `stop` | Stop server (cookies auto-saved) |

## Selector Syntax

| Type | Example | Description |
|------|---------|-------------|
| CSS | `#username`, `input[name=email]` | Standard CSS selector |
| Text | `text=Submit` | Element containing text |
| Role | `role=button` | ARIA role selector |
| XPath | `xpath=//div[@id="main"]` | XPath expression |

## Multi-Session Support

Run isolated browser sessions for parallel agent execution:

```bash
# Each session gets its own browser instance
browser-cli --session agent-1 navigate https://site1.com
browser-cli --session agent-2 navigate https://site2.com

# List all sessions
browser-cli session-list

# Stop specific session
browser-cli --session agent-1 stop
```

## Web Component Support

### `smart-click` — Click Web Components

Web Components often use internal callbacks instead of standard DOM events. `smart-click` auto-detects and calls them:

```bash
browser-cli smart-click "custom-button"
browser-cli smart-click "[data-action=publish]"
```

Detection patterns: `_on*`, `_handle*`, `handle*`, `_click`, `_submit`, `_action`

### `pick` — Discover Element Internals

Inspect elements at specific coordinates and discover their structure:

```bash
browser-cli pick 500 300 --depth=5
```

Returns tag name, selector, detected methods, Shadow DOM structure, and suggestions.

## Dialog Detection

Browser-CLI detects JavaScript dialogs and lets AI decide how to handle them:

```bash
browser-cli dialog-status
# → {"dialog": {"type": "confirm", "message": "Are you sure?"}}

browser-cli dialog-accept     # Accept
browser-cli dialog-dismiss    # Dismiss
```

Supported types: `alert`, `confirm`, `prompt`, `beforeunload`

> Note: Custom HTML popups are not detected — use element selectors instead.

## Output Format

### JSON (recommended for AI agents)

```json
{"command": "navigate", "status": "success", "data": {"url": "https://example.com/", "title": "Example Domain"}}
```

### Markdown (human-readable)

```
## Navigate
- Status: success
- URL: https://example.com/
- Title: Example Domain
```

## AI Integration Guide

### Integration Files Included

| File | Tool | Description |
|------|------|-------------|
| `skills/browser-cli/SKILL.md` | GAL | Full skill definition with patterns |
| `integrations/claude/browser.md` | Claude Code | Custom slash command (copy to `.claude/commands/`) |
| `integrations/codex/browser-cli.md` | OpenAI Codex | Skill reference (copy to `~/.codex/skills/`) |
| `AGENTS.md` | Cursor, Windsurf, etc. | Agent instructions |

### Best Practices for AI Agents

1. **Server auto-starts** — Just run commands, no manual server management
2. **Use `--output json`** — Structured output for reliable parsing
3. **Check `status` field** — Always `"success"` or `"error"`
4. **Handle dialogs** — Check `dialog-status` after clicks that may trigger alerts
5. **Use `--session`** — Isolate parallel agent tasks
6. **Call `stop` when done** — Clean up browser resources
7. **Use `--proxy`** — If behind a firewall, pass proxy URL explicitly
8. **Use `smart-click`** — When normal `click` doesn't work on custom elements

### Example: AI Agent Workflow

```bash
# 1. Navigate and login
browser-cli navigate https://login.example.com
browser-cli fill "#username" "user"
browser-cli fill "#password" "pass"
browser-cli click "button[type=submit]"

# 2. Wait for page load
browser-cli wait ".dashboard"

# 3. Extract data
browser-cli text
browser-cli eval "JSON.stringify(Array.from(document.querySelectorAll('.item')).map(e=>e.textContent))"

# 4. Export
browser-cli screenshot result.png
browser-cli pdf report.pdf

# 5. Cleanup
browser-cli stop
```

## Storage Paths

| Resource | Path |
|----------|------|
| Server socket | `/tmp/browser-cli/server.sock` (single socket for all sessions) |
| Cookies | `/tmp/browser-cli/cookies/<session-id>/<domain>.json` (per session, per domain) |

## Requirements

- Go 1.21+
- Playwright browsers (`make setup-browsers`)

## License

[Apache License 2.0](LICENSE)

Copyright 2024 zmysysz
