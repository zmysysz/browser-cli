# Browser-CLI Platform Promotion Research Report

**Date:** 2026-06-10  
**Project:** [zmysysz/browser-cli](https://github.com/zmysysz/browser-cli) — Go + Playwright browser automation CLI for AI agents

---

## Executive Summary

Browser-cli has **4 existing PRs to awesome-lists** (all still open/pending). The highest-impact opportunities are: **awesome-go** (175K stars), **awesome-ai-agents** (28K stars), and **awesome-browser-automation** (618 stars). For social platforms, **Twitter/X API** and **Medium cross-posting** are the most programmatically feasible. **Product Hunt** requires a manual launch but offers the biggest single-day traffic spike.

---

## Platform-by-Platform Analysis

### 1. 🟧 Product Hunt

| Aspect | Details |
|--------|---------|
| **API** | GraphQL API v2 — requires OAuth2 access token (developer account needed) |
| **Post programmatically?** | ❌ No. Product Hunt requires a **manual launch process** — you schedule a "launch day" and the community votes |
| **Maker requirement** | You **must** be registered as a "Maker" on the product page. Anyone can submit a product, but the maker claim process ties it to your account |
| **Process** | 1) Create PH account 2) Submit product with tagline, description, screenshots, demo video 3) Schedule launch day 4) Engage community on launch day |
| **Effort** | 🔴 High (2-3 hours for assets + launch day engagement) |
| **Reach** | 🟢 Very High (1K–50K visits on launch day if well-received) |
| **Priority** | ⭐⭐⭐⭐ (4/5) — Big traffic spike, but requires preparation and manual work |

**Recommendation:** Prepare a PH launch kit (tagline, 60-sec demo video, 5 screenshots, first comment). Launch on a Tuesday/Wednesday for best visibility. **Cannot be done via API alone.**

---

### 2. 🟩 Indie Hackers

| Aspect | Details |
|--------|---------|
| **API** | No public API. Firebase-backed Ember.js app |
| **Post programmatically?** | ❌ No. Must use the web UI |
| **Showcase format** | They have "Products" (like a directory) and "Posts" (forum threads). You can create a product page and write a launch post |
| **Community** | Active indie founder community, good for early-stage tools |
| **Effort** | 🟡 Medium (30 min to create product + write a post) |
| **Reach** | 🟡 Medium (hundreds to low thousands of targeted developers) |
| **Priority** | ⭐⭐⭐ (3/5) — Good audience fit but manual |

**Recommendation:** Create a product page at indiehackers.com/products and write a post about "Building browser-cli: Giving AI agents browser superpowers." Manual only.

---

### 3. 🟦 Stack Overflow

| Aspect | Details |
|--------|---------|
| **API** | ✅ Stack Exchange API v2.3 — fully accessible (confirmed: 299/300 quota remaining) |
| **Post programmatically?** | ⚠️ Partial — can search for questions via API, but **answering requires OAuth2 + write access token** |
| **Self-promotion rules** | 🔴 **STRICT** — Must disclose affiliation. Answers must directly answer the question; mentioning your tool is allowed only if it's genuinely the best answer. Repeated self-promotion = ban |
| **Relevant questions found** | - [75327403] "OAuth Authorization in Golang CLI Applications" (1.7K views) - [77587298] "Use browser for authentication from console app" (84 views) - [45808799] "How to get HTTP response body using chromedp?" (9.4K views) - [44067030] "How to use Chrome headless with chromedp?" (11K views) |
| **Effort** | 🟡 Medium (crafting high-quality answers takes time) |
| **Reach** | 🟢 High per-answer (evergreen SEO traffic for years) |
| **Priority** | ⭐⭐⭐ (3/5) — Great long-tail SEO, but must be genuinely helpful |

**Recommendation:** Answer 3-5 questions where browser-cli genuinely solves the problem (e.g., "How to automate browser from Go CLI?" type questions). **Always disclose you're the author.** Focus on providing value, not just promotion.

---

### 4. 💬 Discord/Slack Communities

| Aspect | Details |
|--------|---------|
| **API** | Discord API available but requires bot token + server membership |
| **Post programmatically?** | ❌ Not appropriate — communities have anti-spam rules |
| **Key communities to join** | **Discord:** Playwright Official (discord.gg/playwright), Go/Gopher Slack (invite.slack.golangbridge.org), AI Agent communities **Slack:** Gophers (gophers.slack.com), GoBridge |
| **Effort** | 🟢 Low (just join and participate naturally) |
| **Reach** | 🟡 Medium (hundreds of engaged developers per community) |
| **Priority** | ⭐⭐⭐⭐ (4/5) — Low effort, high engagement, builds relationships |

**Recommendation:** Join these communities and participate naturally. When browser-cli is relevant to a conversation, mention it. **Do NOT spam or cold-post.** The awesome-go README itself links to the Gophers Slack.

---

### 5. 🎥 YouTube

| Aspect | Details |
|--------|---------|
| **API** | YouTube Data API v3 — requires OAuth2 + project setup |
| **Post programmatically?** | ⚠️ Can upload videos via API, but video creation is manual |
| **Content ideas** | - 2-min "Browser-CLI in 60 seconds" demo - "Controlling a browser from Claude Code" tutorial - "AI Agent Browser Automation with Go" deep-dive |
| **Effort** | 🔴 High (video recording + editing = 2-4 hours minimum) |
| **Reach** | 🟢 Potentially Very High (YouTube SEO is powerful for dev tools) |
| **Priority** | ⭐⭐ (2/5) — High effort but high ceiling; deprioritize until other channels are done |

**Recommendation:** Create a simple screen-recording demo first (even just a Loom/asciinema recording). A proper YouTube tutorial is a "Phase 2" activity. Consider asciinema.org for a quick terminal recording that can be embedded in README.

---

### 6. 🐦 Twitter/X

| Aspect | Details |
|--------|---------|
| **API** | ✅ Twitter API v2 — POST /2/tweets endpoint available |
| **Post programmatically?** | ✅ **YES** — with a Developer account + Bearer Token + OAuth2 |
| **Cost** | 🔴 **$100/month** for Basic tier (required for posting). Free tier = read-only |
| **Process** | 1) Apply for Developer account 2) Create app 3) Get OAuth2 credentials 4) POST tweets via API |
| **Effort** | 🟡 Medium (API setup + content creation) |
| **Reach** | 🟢 High (Go community, AI agent community, dev tool enthusiasts) |
| **Priority** | ⭐⭐⭐ (3/5) — Feasible but $100/mo is steep for just tweeting |

**Recommendation:** If you already have a Twitter account, just tweet manually. The API is only worth it if you're automating a content pipeline. **Consider scheduling tools (Buffer, Typefully) instead of direct API.**

---

### 7. 💼 LinkedIn

| Aspect | Details |
|--------|---------|
| **API** | LinkedIn Marketing API — requires app approval, restricted access |
| **Post programmatically?** | ⚠️ Possible but requires LinkedIn Developer App + OAuth2 + review process |
| **Groups** | "Go Language", "AI & Machine Learning", "Software Architecture", "DevOps" |
| **Effort** | 🔴 High (API approval takes days; manual posting is easier) |
| **Reach** | 🟡 Medium (LinkedIn dev content gets moderate engagement) |
| **Priority** | ⭐⭐ (2/5) — Low ROI for dev tools |

**Recommendation:** Post manually on your personal LinkedIn feed and in relevant groups. Don't bother with the API. Write a short article-format post about the problem browser-cli solves.

---

### 8. 📝 Medium

| Aspect | Details |
|--------|---------|
| **API** | ✅ Medium API v1 — `api.medium.com/v1` (OAuth2 based) |
| **Post programmatically?** | ✅ **YES** — POST `/v1/users/{userId}/posts` with Integration Token |
| **Process** | 1) Get Medium Integration Token from Settings → Integration Tokens 2) GET `/v1/me` for userId 3) POST article with HTML/Markdown content |
| **Cross-posting** | ✅ Can republish Dev.to article (use canonical URL to preserve SEO) |
| **Effort** | 🟢 Low (if article already exists, just reformat and post) |
| **Reach** | 🟡 Medium-High (Medium has good SEO and algorithmic distribution) |
| **Priority** | ⭐⭐⭐⭐ (4/5) — Easy cross-post, good SEO, programmatically feasible |

**Recommendation:** **Do this.** Get an Integration Token, adapt the Dev.to article, and post with `canonicalUrl` pointing to the original. This is the best effort-to-reach ratio for content.

---

## Awesome Lists — Submission Status & Targets

### Already Submitted (All Open/Pending)

| List | Stars | PR Status | Section |
|------|-------|-----------|---------|
| [lorien/awesome-web-scraping](https://github.com/lorien/awesome-web-scraping) | 7,923 | [#249](https://github.com/lorien/awesome-web-scraping/pull/249) — Open | CLI tools |
| [agarrharr/awesome-cli-apps](https://github.com/agarrharr/awesome-cli-apps) | ~15K | [#1148](https://github.com/agarrharr/awesome-cli-apps/pull/1148) — Open | Browser Replacement |
| [e2b-dev/awesome-ai-agents](https://github.com/e2b-dev/awesome-ai-agents) | 28,229 | [#1069](https://github.com/e2b-dev/awesome-ai-agents/pull/1069) — Open | AI agents |
| [mxschmitt/awesome-playwright](https://github.com/mxschmitt/awesome-playwright) | 1,488 | [#153](https://github.com/mxschmitt/awesome-playwright/pull/153) — Open | AI & Agents |

### High-Priority Targets (Not Yet Submitted)

| List | Stars | Where browser-cli fits | Priority |
|------|-------|----------------------|----------|
| [avelino/awesome-go](https://github.com/avelino/awesome-go) | **175,058** | "Selenium and browser control tools" section (alongside chromedp, rod, playwright-go) | 🔴 **CRITICAL** |
| [angrykoala/awesome-browser-automation](https://github.com/angrykoala/awesome-browser-automation) | 618 | Main list (alongside Playwright, Puppeteer, Chromedp) | ⭐⭐⭐⭐⭐ |
| [transitive-bullshit/awesome-puppeteer](https://github.com/transitive-bullshit/awesome-puppeteer) | 2,557 | Related tools / alternatives section | ⭐⭐⭐ |
| [Shubhamsaboo/awesome-llm-apps](https://github.com/Shubhamsaboo/awesome-llm-apps) | 114,106 | Agent tools section | ⭐⭐⭐⭐⭐ |
| [jim-schwoebel/awesome_ai_agents](https://github.com/jim-schwoebel/awesome_ai_agents) | 1,810 | Agent tools | ⭐⭐⭐ |
| [h4ckf0r0day/awesome-ai-web-scraping](https://github.com/h4ckf0r0day/awesome-ai-web-scraping) | 52 | AI + scraping tools | ⭐⭐ |

### Medium-Priority Targets

| List | Stars | Fit |
|------|-------|-----|
| [mantcz/awesome-go-cli](https://github.com/mantcz/awesome-go-cli) | 37 | Go CLI tools |
| [awesome-directories/cli](https://github.com/awesome-directories/cli) | 1 | SaaS directory submission tool |

---

## Priority Ranking Summary

| Rank | Action | Effort | Reach | Feasibility |
|------|--------|--------|-------|-------------|
| **1** | 🟢 Submit PR to **awesome-go** (175K stars) | Low | **Massive** | Easy — just add 1 line to README |
| **2** | 🟢 Submit PR to **awesome-browser-automation** | Low | High | Easy — natural fit alongside chromedp |
| **3** | 🟢 Cross-post to **Medium** via API | Low | Medium-High | Easy — token-based posting |
| **4** | 🟡 Join **Discord/Slack** communities | Low | Medium | Manual — just sign up and engage |
| **5** | 🟡 Follow up on **4 existing PRs** | Low | High | Nudge maintainers politely |
| **6** | 🟡 Submit PR to **awesome-llm-apps** (114K stars) | Low | Massive | Easy — natural fit for AI agent tools |
| **7** | 🟡 Answer **Stack Overflow** questions | Medium | High (long-tail) | Semi-automated search |
| **8** | 🟠 Prepare **Product Hunt** launch | High | Very High (1 day) | Manual launch day |
| **9** | 🟠 Post on **Indie Hackers** | Medium | Medium | Manual web UI |
| **10** | 🔴 Create **YouTube** demo video | High | Potentially High | Manual recording |
| **11** | 🔴 **Twitter/X** API posting | Medium | High | $100/mo — manual is free |
| **12** | 🔴 **LinkedIn** API posting | High | Medium | Not worth API hassle |

---

## Immediate Next Steps (Actionable)

### Phase 1 — This Week (Low effort, high impact)
1. **Submit PR to awesome-go** — Add browser-cli to "Selenium and browser control tools" section. This is the single highest-impact action possible (175K stars).
2. **Submit PR to awesome-browser-automation** — Natural fit alongside chromedp and Playwright entries.
3. **Submit PR to awesome-llm-apps** (114K stars) — Add to AI agent tools section.
4. **Follow up on 4 existing PRs** — Comment politely on the open PRs to get maintainer attention.
5. **Cross-post Dev.to article to Medium** — Get Integration Token, POST via API.

### Phase 2 — Next 2 Weeks (Medium effort)
6. **Join Discord communities** — Playwright, Gophers Slack, AI agent servers. Engage naturally.
7. **Answer 3-5 Stack Overflow questions** — Search for "go browser automation CLI" questions, provide genuinely helpful answers mentioning browser-cli.
8. **Post on Indie Hackers** — Create product page + write a post.

### Phase 3 — When Ready (Higher effort)
9. **Prepare Product Hunt launch** — Create demo video, screenshots, schedule launch.
10. **Create YouTube demo** — Even a 2-minute screen recording goes a long way.

---

## Technical Notes

### Medium API Posting (Feasible Now)
```bash
# 1. Get Integration Token from: https://medium.com/me/settings/security
# 2. Get user ID:
curl -H "Authorization: Bearer YOUR_TOKEN" https://api.medium.com/v1/me

# 3. Post article:
curl -X POST https://api.medium.com/v1/users/{userId}/posts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Give Your AI Agent Browser Superpowers with browser-cli",
    "contentFormat": "markdown",
    "content": "<article markdown content>",
    "canonicalUrl": "https://dev.to/zmysysz/browser-cli-...",
    "tags": ["go", "playwright", "ai", "automation", "cli"],
    "publishStatus": "public"
  }'
```

### awesome-go PR (Ready to Submit)
Add this line under `### Selenium and browser control tools`:
```markdown
- [browser-cli](https://github.com/zmysysz/browser-cli) - CLI for browser automation designed for AI agents, built with Go and Playwright.
```

### awesome-browser-automation PR (Ready to Submit)
Add this line under the Go section:
```markdown
* [browser-cli](https://github.com/zmysysz/browser-cli) - CLI tool for browser automation designed for AI agents. Built with Go and Playwright.
```
