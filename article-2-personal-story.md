# Article 2: Personal Story / Builder Angle (for cross-posting)

---

title: "I built a CLI tool so my AI coding agent can browse the web"
published: false
description: "My AI assistant could write code but couldn't test it in a browser. So I built browser-cli — and it changed how I work."
tags: javascript, python, opensource, productivity, ai, sideproject, webdev
cover_image: https://github.com/zmysysz/browser-cli/raw/main/docs/demo.gif

---

# I Built a CLI Tool So My AI Coding Agent Can Browse the Web

Here's a problem I kept running into:

I'd ask Claude Code to build a feature. It writes the code perfectly. Then I'd say, *"Now go test it — open the app, fill in the form, submit it, and tell me if it works."*

And Claude would say: *"I can't interact with browsers."*

Every. Single. Time.

My AI assistant could write a full React component from scratch, but couldn't click a button on a webpage. It could generate a Playwright test script, but couldn't *run* it. The intelligence was there — the hands weren't.

So I built the hands.

## The Idea Was Simple

What if browser automation was just... shell commands?

```bash
browser-cli navigate https://my-app.com
browser-cli fill "#email" "test@example.com"
browser-cli click "button[type=submit]"
browser-cli text
```

No scripts. No Node.js setup. No Python virtualenv. Just commands that any AI agent can run as a shell command, with JSON output it can actually parse.

I called it **browser-cli**.

## What It Does

Browser-CLI wraps Playwright (the browser automation engine) into a Go CLI tool. The architecture is straightforward:

1. A background server manages browser instances over Unix sockets
2. CLI commands connect, execute, and return JSON
3. The server auto-starts when you run your first command — no manual setup

Every command returns structured JSON:

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

This means AI agents can call browser-cli, read the response, and decide what to do next. It's a feedback loop: navigate → read → decide → act.

## The Features I Actually Needed

I didn't start with a feature list. I started with my own frustrations and built what I needed:

**Login persistence.** I got tired of logging into GitHub every time I wanted to scrape something. So I added a `login` command that opens a visible browser, lets you log in manually, and saves the entire state (cookies + localStorage) to a file. Next time, you just pass `--state ./github-state.json` and you're already logged in.

```bash
browser-cli login https://github.com
# ... log in manually, press Ctrl+C when done ...

# Reuse forever
browser-cli --state ./github-state.json navigate https://github.com/settings
```

**Stealth mode.** Half the sites I tried to automate blocked Playwright immediately. So I built in anti-detection: override `navigator.webdriver`, fake plugins and mimeTypes, add `window.chrome`, disable automation flags. Now it works on Google sign-in, Cloudflare-protected sites, and most bot detectors.

**Session isolation.** Sometimes I'd have two AI agents working on different tasks simultaneously, and they'd interfere with each other's browser state. The `--session` flag gives each agent its own isolated browser instance:

```bash
browser-cli --session agent-1 navigate https://site-a.com
browser-cli --session agent-2 navigate https://site-b.com
```

**Web Components.** Modern SPAs use custom elements that ignore standard DOM clicks. I added `smart-click` that detects internal handlers (`_onClick`, `handleSubmit`, etc.) and calls them directly.

## How I Use It Day to Day

Here's my actual workflow with Claude Code:

1. I ask Claude to build a feature
2. Claude writes the code
3. I say: *"Test it with browser-cli"*
4. Claude runs `browser-cli navigate http://localhost:3000`, fills in forms, clicks buttons, reads the output
5. If something's broken, Claude sees the error and fixes it — without me touching the browser

It's a tight loop. No context switching. No "let me check that manually and get back to you."

Another thing I do: scraping. I needed to pull data from a dashboard that has no API. Instead of writing a one-off Python script, I just tell Claude: *"Go to this dashboard, extract the numbers from the table, and put them in a CSV."* Claude uses browser-cli to navigate, extract, and format the data. Done.

## The Tech Stack

- **Go** — I wanted a single static binary with no runtime dependencies. Go is perfect for this. `CGO_ENABLED=0` and it runs anywhere.
- **Playwright** — The best browser automation engine. Handles Chromium, Firefox, and WebKit.
- **JSON-RPC over Unix sockets** — Simple, fast, no HTTP overhead. The server manages browser instances and the CLI is just a thin client.

The whole thing compiles to one binary. No Node.js. No Python. No runtime dependencies.

## It's Open Source

I put it on GitHub because I figured other people have the same problem:

👉 **https://github.com/zmysysz/browser-cli**

Apache 2.0 license. Free to use, modify, and distribute.

```bash
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build
make setup-browsers
browser-cli navigate https://example.com
```

Integration files are included for Claude Code, OpenAI Codex, Cursor, and generic AI agents. Just copy the right file to the right directory and you're set.

## What's Next

I'm working on:
- **CDP endpoint support** — Connect to an externally-launched Chrome for sites with strict bot detection
- **Better element discovery** — Making it easier for AI agents to find the right selectors
- **More integrations** — Windsurf, Aider, and other AI coding tools

If you try it, I'd love to hear what you think. Open an issue, drop a star, or just tell me what workflow you'd automate first.

---

*What would you do if your AI coding assistant could browse the web?*
