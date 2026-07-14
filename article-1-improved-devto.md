# Article 1: Improved Dev.to Article

---

title: "How to Give Your AI Agent Browser Superpowers with Browser-CLI"
published: false
description: "A CLI tool that lets Claude Code, Cursor, and Codex browse the web, fill forms, and scrape data — no scripts required."
tags: javascript, python, opensource, productivity, ai, webdev, automation, cli
cover_image: https://github.com/zmysysz/browser-cli/raw/main/docs/demo.gif

---

# How to Give Your AI Agent Browser Superpowers with Browser-CLI

**Your AI coding assistant just wrote perfect code. Now you need it to test the login flow on your staging site.**

You ask Claude Code: *"Go to staging.example.com, log in with test@example.com, and verify the dashboard loads."*

Claude's response: *"I can't interact with web browsers. I can write Playwright scripts for you, but I can't run them."*

**This is the missing piece.** Your AI agent has all the intelligence to navigate websites, fill forms, and extract data — but no hands to actually do it.

Browser-CLI fixes this. It's a command-line tool that gives AI agents direct browser control through simple shell commands. No scripts. No complex setup. Just:

```bash
browser-cli navigate https://staging.example.com
browser-cli fill "#email" "test@example.com"
browser-cli fill "#password" "secret"
browser-cli click "button[type=submit]"
browser-cli wait ".dashboard"
browser-cli screenshot result.png
```

Every command returns structured JSON that AI agents can parse and act on. The result? Your AI agent can now browse the web, test your apps, and automate repetitive tasks — all by itself.

---

## Why You Need This

If you use any AI coding assistant (Claude Code, Cursor, Codex, Windsurf, etc.), you've probably hit these walls:

| Task | AI Agent Can Do It? | Why Not? |
|------|---------------------|----------|
| Log into a website and scrape data | ❌ | No browser access |
| Fill out and submit forms | ❌ | No browser access |
| Take screenshots of pages | ❌ | No browser access |
| Test web apps end-to-end | ❌ | No browser access |
| Automate repetitive web tasks | ❌ | No browser access |

**Existing solutions don't fit AI agents:**

- **Puppeteer/Playwright**: Require writing JavaScript/Python scripts. AI agents can generate these scripts, but running them requires a Node.js/Python environment, proper setup, and error handling. It's fragile and slow.
- **browser-use**: A Python library that uses LLMs to control browsers. But it's designed for Python developers, not as a tool for AI coding assistants to call directly.
- **Actionbook**: Similar concept, but requires specific integrations and doesn't have the AI-first design.

**Browser-CLI is different.** It's designed from the ground up for AI agents to call as a shell command:

- ✅ **Zero-code**: Just CLI commands, no scripts to write or maintain
- ✅ **JSON output**: Structured responses that AI agents can parse reliably
- ✅ **Auto-managed server**: No manual server start/stop, it just works
- ✅ **Session isolation**: Multiple agents can run in parallel without conflicts
- ✅ **Login persistence**: Log in once manually, reuse forever in automation
- ✅ **Stealth mode**: Bypasses bot detection on sites like Google, Cloudflare

---

## Quick Start: 5 Minutes to Browser Superpowers

### Step 1: Install

```bash
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build
make setup-browsers  # Downloads Playwright browsers (first time only)
make install         # Adds to PATH
```

**Requirements:** Go 1.21+ (for building). The binary is fully static — no CGO, no runtime dependencies.

### Step 2: Try It

```bash
# Navigate to a page (server auto-starts)
browser-cli navigate https://example.com

# Extract page content
browser-cli text
```

**Output:**
```json
{
  "command": "text",
  "status": "success",
  "data": {
    "content": "Example Domain\nThis domain is for use in illustrative examples...",
    "url": "https://example.com/",
    "title": "Example Domain"
  }
}
```

### Step 3: Integrate with Your AI Tool

**Claude Code:**
```bash
mkdir -p .claude/commands/
cp integrations/claude/browser.md .claude/commands/
```

**OpenAI Codex:**
```bash
mkdir -p ~/.codex/skills/
cp integrations/codex/browser-cli.md ~/.codex/skills/
```

**Cursor / Windsurf / Others:** The `AGENTS.md` file at the project root is automatically read by most AI coding tools.

Now just ask your AI: *"Use browser-cli to check the top 5 posts on Hacker News"*

---

## Real-World Examples

### Example 1: Automated Web Scraping

```bash
# Navigate to Hacker News
browser-cli navigate https://news.ycombinator.com

# Extract the top 5 post titles using JavaScript
browser-cli eval "JSON.stringify(
  Array.from(document.querySelectorAll('.titleline > a'))
    .slice(0, 5)
    .map(a => ({title: a.textContent, link: a.href}))
)"
```

**Output:**
```json
{
  "command": "eval",
  "status": "success",
  "data": {
    "result": "[{\"title\":\"Show HN: I built a CLI tool for AI agents\",\"link\":\"https://github.com/...\"}, ...]"
  }
}
```

### Example 2: Login and Scrape Protected Content

```bash
# First time: log in manually (opens a visible browser)
browser-cli login https://github.com
# ... complete login in the browser window ...
# Press Ctrl+C when done — state is saved to ./github-state.json

# Now automate using the saved login state
browser-cli --state ./github-state.json navigate https://github.com/settings
browser-cli text
```

**The login state persists forever.** You can reuse it across sessions, projects, and even different AI agents.

### Example 3: Form Submission

```bash
browser-cli navigate https://httpbin.org/forms/post

# Fill multiple fields
browser-cli fill "input[name='custname']" "John Doe"
browser-cli fill "input[name='custtel']" "555-1234"
browser-cli fill "input[name='custemail']" "john@example.com"
browser-cli fill "textarea[name='comments']" "This is a test submission"

# Select radio button
browser-cli click "input[value='medium']"

# Submit
browser-cli click "button[type=submit]"

# Get the result
browser-cli text
```

### Example 4: Multi-Step Workflow in One Command

```bash
browser-cli run "navigate https://example.com; click a; text; screenshot result.png"
```

The `run` command executes multiple actions in a single round-trip, reducing latency for complex workflows.

### Example 5: Handle JavaScript Dialogs

```bash
browser-cli navigate https://example.com/delete
browser-cli click ".delete-button"

# Check if a dialog appeared
browser-cli dialog-status
# → {"dialog": {"type": "confirm", "message": "Are you sure?"}}

# Accept or dismiss
browser-cli dialog-accept
# or: browser-cli dialog-dismiss
```

---

## Feature Deep Dive

### 🥷 Stealth Mode: Bypass Bot Detection

Browser-CLI includes built-in stealth mode that overrides automation fingerprints:

- **`navigator.webdriver`** — Set to `undefined` instead of `true`
- **`navigator.plugins`** — Faked with realistic entries
- **`navigator.mimeTypes`** — Includes PDF viewer
- **`window.chrome`** — Added to match real Chrome
- **Permissions API** — Auto-grants notifications
- **Chromium flags** — Disables `AutomationControlled` blink feature

This lets you automate sites that normally block bots, like Google sign-in, Cloudflare-protected pages, and banking sites.

```bash
# Use system Chrome for even better stealth
browser-cli --chrome navigate https://accounts.google.com
```

### 🍪 Login Persistence: Set and Forget

The `login` command opens a visible browser for you to log in manually. Once done, the entire browser state (cookies + localStorage) is saved to a JSON file:

```bash
# Log in once
browser-cli login https://twitter.com
# → saves to ./twitter-login.json

# Reuse forever
browser-cli --state ./twitter-login.json navigate https://twitter.com/home
browser-cli text  # You're already logged in!
```

### 🔗 Session Isolation: Parallel Agents

Multiple AI agents can run simultaneously without interfering with each other:

```bash
# Agent 1 works on site A
browser-cli --session agent-1 navigate https://site-a.com

# Agent 2 works on site B (completely isolated)
browser-cli --session agent-2 navigate https://site-b.com

# List all active sessions
browser-cli session-list
```

Each session gets its own browser instance, cookies, and storage.

### 🎯 Web Components Support

Modern web apps use custom elements that don't respond to standard clicks. Browser-CLI's `smart-click` detects and triggers internal handlers:

```bash
# Normal click might not work on custom elements
browser-cli click "my-custom-button"  # ❌ May fail

# Smart-click finds and calls internal methods
browser-cli smart-click "my-custom-button"  # ✅ Works!
```

Detection patterns: `_on*`, `_handle*`, `handle*`, `_click`, `_submit`, `_action`

### 📸 Screenshots, PDFs, and More

```bash
# Screenshot
browser-cli screenshot page.png

# Full page PDF (Chromium only)
browser-cli pdf report.pdf

# Screenshot specific element
browser-cli elements ".chart"
browser-cli eval "document.querySelector('.chart').scrollIntoView()"
browser-cli screenshot chart.png
```

---

## Comparison with Alternatives

| Feature | Browser-CLI | Puppeteer | Playwright | browser-use | Actionbook |
|---------|-------------|-----------|------------|-------------|------------|
| **AI-friendly output** | JSON by default | JS API | JS/Python API | Natural language | JSON |
| **Zero-code usage** | ✅ CLI commands | ❌ Scripts | ❌ Scripts | ✅ LLM-driven | ✅ Actions |
| **Session isolation** | ✅ Built-in | ❌ Manual | ❌ Manual | ❌ | ✅ |
| **Login persistence** | ✅ Built-in | ❌ Manual | ❌ Manual | ❌ | ✅ |
| **Stealth mode** | ✅ Built-in | ❌ Plugin | ❌ Plugin | ✅ | ❌ |
| **Web Components** | ✅ smart-click | ❌ | ❌ | ✅ | ❌ |
| **Language** | Go (single binary) | Node.js | Node/Python | Python | Go |
| **Runtime deps** | None | Node.js | Node/Python | Python + deps | None |
| **Setup complexity** | Low | Medium | Medium | High | Low |

**When to use Browser-CLI:**
- You want AI agents to control browsers directly
- You need a simple, script-free automation solution
- You want login state persistence
- You need to bypass bot detection
- You want parallel agent execution

**When to use Puppeteer/Playwright:**
- You're writing traditional test scripts
- You need fine-grained control over browser behavior
- You're building a testing framework

**When to use browser-use:**
- You want natural language browser control
- You're building an LLM-powered application

---

## All Commands Reference

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
| `smart-click <selector>` | Click Web Components |
| `fill <selector> <value>` | Fill input field |
| `type <selector> <text>` | Type character by character |
| `select <selector> <value>` | Select dropdown option |
| `hover <selector>` | Hover over element |
| `keyboard <key>` | Press key (e.g., `Ctrl+A`, `Enter`) |
| `upload <selector> <file>` | Upload file |
| `eval <script>` | Execute JavaScript |

### Extraction
| Command | Description |
|---------|-------------|
| `text` | Extract page text |
| `screenshot [path]` | Take screenshot |
| `elements <selector>` | Find elements |
| `pdf [file]` | Save as PDF |

### Utility
| Command | Description |
|---------|-------------|
| `wait <selector>` | Wait for element |
| `scroll <up\|down>` | Scroll page |
| `pick <x> <y>` | Inspect element at coordinates |

### Tabs & Sessions
| Command | Description |
|---------|-------------|
| `tab-new` | Create new tab |
| `tab-list` | List all tabs |
| `tab-switch <id>` | Switch to tab |
| `session-list` | List active sessions |
| `stop` | Stop server |

---

## Call to Action

**Star the repo:** https://github.com/zmysysz/browser-cli ⭐

**Try it now:**
```bash
git clone https://github.com/zmysysz/browser-cli
cd browser-cli && make build && make setup-browsers
browser-cli navigate https://example.com
browser-cli text
```

**Integrate with your AI tool:**
- Claude Code: Copy `integrations/claude/browser.md` to `.claude/commands/`
- Codex: Copy `integrations/codex/browser-cli.md` to `~/.codex/skills/`
- Others: Check the `integrations/` folder

**Have questions or feature requests?** Open an issue on GitHub or drop a comment below!

---

*Browser-CLI is open source under Apache 2.0. Built with Go + Playwright. No CGO required — works on Linux, macOS, and Windows.*
