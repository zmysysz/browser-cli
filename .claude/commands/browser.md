# Browser Automation with Browser-CLI

Use browser-cli to automate browser interactions. The server starts automatically.

## Quick Reference

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
browser-cli keyboard <key>           # e.g. "Ctrl+A", "Enter", "Tab"
browser-cli upload <selector> <file>
browser-cli smart-click <selector>   # for Web Components

# Extraction
browser-cli text                     # page text content
browser-cli screenshot [path]
browser-cli elements <selector>
browser-cli eval <javascript>
browser-cli pdf [file]               # Chromium only

# Utility
browser-cli wait <selector>
browser-cli scroll up|down
browser-cli pick <x> <y> [--depth=N]

# Tabs
browser-cli tab-new | tab-list | tab-switch <id> | tab-close [id]

# Dialogs
browser-cli dialog-status | dialog-accept [value] | dialog-dismiss

# Sessions & Cleanup
browser-cli --session <id> <cmd>     # isolated session
browser-cli stop                     # stop server, save cookies
```

## Selector Types
- CSS: `#id`, `.class`, `button[type=submit]`
- Text: `text=Submit`
- Role: `role=button`
- XPath: `xpath=//div[@id="main"]`

## Key Flags
- `--output json` — structured output for parsing
- `--session <id>` — isolated browser instance
- `--proxy http://host:port` — proxy for network access
- `--headless=false` — show browser window (debugging)
- `--timeout 60s` — operation timeout

## Patterns

### Login
```bash
browser-cli navigate https://site.com/login
browser-cli fill "#email" "user@example.com"
browser-cli fill "#password" "secret"
browser-cli click "button[type=submit]"
browser-cli wait ".dashboard"
```

### Extract data
```bash
browser-cli navigate https://site.com/data
browser-cli eval "JSON.stringify(Array.from(document.querySelectorAll('.item')).map(e=>e.textContent))"
```

### Handle dialogs
```bash
browser-cli click ".delete-btn"
browser-cli dialog-status
browser-cli dialog-accept
```

Always call `browser-cli stop` when done.
