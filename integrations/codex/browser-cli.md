# Browser-CLI

Browser automation tool for AI agents. Server auto-starts on first command.

## Commands

| Category | Commands |
|----------|----------|
| Navigate | `navigate <url>`, `back`, `forward`, `reload` |
| Click | `click <sel>`, `click-js <sel>`, `smart-click <sel>`, `right-click <sel>`, `dblclick <sel>` |
| Input | `fill <sel> <val>`, `type <sel> <text>`, `select <sel> <val>`, `keyboard <key>`, `upload <sel> <file>` |
| Extract | `text`, `screenshot [path]`, `elements <sel>`, `eval <js>`, `pdf [file]` |
| Utility | `wait <sel>`, `scroll up\|down`, `pick <x> <y>` |
| Tabs | `tab-new`, `tab-list`, `tab-switch <id>`, `tab-close [id]` |
| Dialogs | `dialog-status`, `dialog-accept [val]`, `dialog-dismiss` |
| Server | `status`, `stop`, `session-list`, `--session <id>` |

## Selectors
CSS (`#id`, `.class`), Text (`text=Submit`), Role (`role=button`), XPath (`xpath=//div`)

## Key Flags
`--output json`, `--session <id>`, `--proxy <url>`, `--headless=false`, `--timeout 60s`

## Usage Pattern
```bash
browser-cli navigate <url>
browser-cli fill "#email" "value"
browser-cli click "button"
browser-cli text
browser-cli stop
```

Always `stop` when done. Cookies auto-persist.
