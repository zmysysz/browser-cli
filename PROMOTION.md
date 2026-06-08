# Browser-CLI 推广材料汇总

## 一、V2EX 发帖

### 节点
分享创造

### 标题
Browser-CLI - 给 AI Agent 用的浏览器自动化 CLI，一条命令操控浏览器

### 正文
大家好，我开发了一个专门给 AI Agent 用的浏览器自动化工具：**Browser-CLI**。

### 背景
现在 AI 编程工具越来越强（Claude Code、Cursor、Codex），但它们缺少一个能力：**操作浏览器**。

想让它帮你：
- 自动登录网站抓数据？❌ 做不到
- 自动填表提交？❌ 做不到
- 自动截图保存？❌ 做不到

Puppeteer/Playwright 需要写 JS/Python 脚本，AI 很难直接调用。

### 解决方案
Browser-CLI 把浏览器能力封装成 CLI，AI 直接调用：

```bash
browser-cli navigate https://example.com
browser-cli fill "#search" "browser automation"
browser-cli click "button[type=submit]"
browser-cli text
browser-cli screenshot result.png
```

### 特点
- **AI 原生**：JSON 输出，命令语义清晰，AI agent 直接调用
- **零代码**：不需要写脚本，一条命令搞定
- **会话隔离**：多 agent 并行不冲突
- **登录持久化**：手动登录一次，自动化复用
- **Web Components 支持**：smart-click 自动检测内部方法

### 使用场景
- AI 自动抓取网页数据
- AI 自动填表、提交
- AI 自动测试 Web 应用
- AI 自动截图、生成报告

### 安装
```bash
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build
make setup-browsers
```

### GitHub
https://github.com/zmysysz/browser-cli

欢迎试用反馈！

---

## 二、awesome-browser-automation PR

### 仓库
https://github.com/angrykoala/awesome-browser-automation

### PR 标题
Add Browser-CLI - AI-first browser automation CLI

### PR 描述
Add Browser-CLI to the AI section. It's a command-line tool designed specifically for AI agents (Claude Code, Cursor, Codex) to control browsers.

Key features:
- AI-first design: JSON output, clear command semantics
- Zero-code: No JS/Python scripts needed, just CLI commands
- Session isolation: Multiple agents can run in parallel
- Login persistence: Manual login once, reuse for automation
- Web Components support: smart-click for custom elements

### README.md 修改
在 `### AI` 部分，按字母顺序在 BrowserBook 和 Browser-Use 之间添加：

```markdown
* [Browser-CLI](https://github.com/zmysysz/browser-cli) - Command-line tool for browser automation designed for AI agents. Built with Go + Playwright, provides JSON output and session isolation for multi-agent workflows.
```

### 操作步骤
1. Fork https://github.com/angrykoala/awesome-browser-automation
2. 在 fork 的 README.md 的 AI 部分添加上述条目
3. 提交 PR

---

## 三、awesome-ai-agents PR（可选）

### 仓库
https://github.com/e2b-dev/awesome-ai-agents

这个列表有表单提交方式，更简单：
https://forms.gle/UXQFCogLYrPFvfoUA

填写内容：
- Product name: Browser-CLI
- GitHub: https://github.com/zmysysz/browser-cli
- Category: Open source / Build your own
- Description: Command-line tool for browser automation designed for AI agents. Built with Go + Playwright.

---

## 四、GitHub Topics

在你的 repo 页面 https://github.com/zmysysz/browser-cli 点击 ⚙️ Add topics，添加：

ai-agent browser-automation playwright cli go golang claude-code cursor codex developer-tools

---

## 五、Twitter/X 推文

给 AI Agent 加上浏览器超能力 🚀

Browser-CLI: 一条命令让 Claude/Codex/Cursor 自动操作浏览器

✅ JSON 输出，AI 友好
✅ 会话隔离，多 agent 并行
✅ 登录状态持久化
✅ Go 单二进制，零依赖

GitHub: https://github.com/zmysysz/browser-cli

#AI #AIAgent #BrowserAutomation #ClaudeCode #Cursor #DevTools

---

## 六、Reddit（需要代理）

### 目标社区
- r/LocalLLaMA（AI agent 工具讨论）
- r/ChatGPTCoding（AI 编程工具）
- r/ClaudeAI（Claude 生态）
- r/webdev（Web 开发工具）

### 帖子标题
Browser-CLI: A CLI tool that gives AI agents browser superpowers (Go + Playwright)

### 帖子内容
Hey everyone! I built Browser-CLI, a command-line tool that lets AI coding assistants (Claude Code, Cursor, Codex) control browsers directly.

Instead of writing Puppeteer/Playwright scripts, AI agents can just run:

```bash
browser-cli navigate https://example.com
browser-cli fill "#search" "query"
browser-cli click "button[type=submit]"
browser-cli text  # Extract page content as JSON
```

Key features:
- **AI-first**: JSON output, clear command semantics
- **Zero-code**: No scripts needed, just CLI commands
- **Session isolation**: Multiple agents can run in parallel
- **Login persistence**: Login once manually, reuse for automation
- **Web Components**: smart-click for custom elements

Built with Go + Playwright. Single binary, no CGO required.

GitHub: https://github.com/zmysysz/browser-cli

Would love your feedback!

---

## 七、Dev.to 博客文章

### 标题
How to Give Your AI Agent Browser Superpowers with Browser-CLI

### 大纲
1. The Problem: AI agents can't use browsers
2. Existing solutions and their limitations (Puppeteer scripts, browser-use, etc.)
3. Introducing Browser-CLI
4. Quick start guide
5. Integration with Claude Code / Cursor / Codex
6. Advanced features (session isolation, login persistence, Web Components)
7. Comparison with alternatives
8. Call to action

---

## 推荐执行顺序

1. ⭐ **GitHub Topics**（1分钟，立即做）
2. ⭐ **V2EX 发帖**（5分钟，需要先登录）
3. ⭐ **awesome-browser-automation PR**（10分钟，fork + edit + PR）
4. **awesome-ai-agents 表单**（2分钟，填表提交）
5. **Twitter 推文**（需要账号）
6. **Reddit 发帖**（需要代理）
7. **Dev.to 博客**（30分钟写文章）
