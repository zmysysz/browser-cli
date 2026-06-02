# Browser-CLI

[English](README.md) | [中文](README_zh.md)

基于 Go + Playwright 的浏览器自动化命令行工具，专为 AI Agent 设计。

**一条命令控制浏览器 — 完美适配 Codex、Claude Code、Cursor 等所有 AI 编程助手。**

```bash
browser-cli navigate https://example.com
browser-cli fill "#search" "browser automation"
browser-cli click "button[type=submit]"
browser-cli text
```

## 特性

- 🤖 **AI 优先** — 结构化 JSON 输出，清晰的命令语义，服务器自动管理
- 🔒 **会话隔离** — 通过 `--session` 为每个 Agent 分配独立浏览器实例
- 🍪 **Cookie 持久化** — 自动保存/加载，跨会话保持登录状态
- 🌐 **代理支持** — `--proxy http://host:port` 突破网络限制
- 🎯 **Web Components** — `smart-click` 和 `pick` 处理自定义元素和 Shadow DOM
- ⌨️ **完整键盘** — 快捷键、组合键、Tab/Enter/Escape、Ctrl+A/C/V
- 📄 **PDF 和截图** — 导出页面为 PDF 或 PNG
- 📁 **文件上传** — 上传文件到任意 `<input type="file">`

## 快速安装

```bash
# 克隆并构建（需要 Go 1.21+）
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build

# 安装 Playwright 浏览器（仅首次需要）
make setup-browsers

# 添加到 PATH
make install

# 或构建无 CGO 的纯静态二进制
make build-static
```

> **无需 CGO。** Browser-CLI 可用 `CGO_ENABLED=0` 编译，支持交叉编译：
> ```bash
> CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/browser-cli.exe .
> CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/browser-cli-mac .
> ```

## AI Agent 30 秒接入

### Claude Code

将项目中的 `.claude/commands/browser.md` 复制到你的项目：

```bash
cp -r .claude/ ~/.claude/   # 项目级别
```

然后在 Claude Code 中直接说：*"Navigate to github.com and take a screenshot"*

### OpenAI Codex

将项目中的 skill 文件复制：

```bash
cp -r .codex/ ~/.codex/
```

### Cursor / Windsurf / 任意 Agent

`AGENTS.md` 已放在项目根目录，大多数 AI 编程工具会自动读取。

### GAL（Global Agent Layer）

```bash
cp -r skills/browser-cli/ ~/.gal/skills/
```

> 📁 所有集成文件已包含在仓库中：`skills/`、`.claude/`、`.codex/`、`AGENTS.md`

## 快速开始

浏览器服务器自动启动 — 直接运行命令即可：

```bash
# 导航（服务器自动启动）
browser-cli navigate https://example.com

# 交互
browser-cli fill "#username" "user"
browser-cli click "button[type=submit]"

# 提取
browser-cli text
browser-cli screenshot page.png

# 完成
browser-cli stop
```

### 单条命令执行多步操作

```bash
browser-cli run "navigate https://example.com; click a; text"
```

## 所有命令

### 全局选项

```
--browser, -b     浏览器：chromium, firefox, webkit（默认：chromium）
--headless        无头模式（默认：true）
--timeout, -t     超时时间（默认：30s）
--output, -o      输出格式：json, markdown（默认：markdown）
--session, -s     隔离浏览器实例的会话 ID
--proxy           代理服务器 URL（如 http://proxy:8080 或 socks5://proxy:1080）
--idle-timeout    空闲自动关闭时间（默认：1h，0 为禁用）
```

### 导航

| 命令 | 说明 |
|------|------|
| `navigate <url>` | 导航到 URL |
| `back` | 后退 |
| `forward` | 前进 |
| `reload` | 刷新页面 |

### 交互

| 命令 | 说明 |
|------|------|
| `click <selector>` | 点击元素 |
| `click-js <selector>` | JS 点击（绕过可见性检查） |
| `smart-click <selector>` | 智能点击 Web Components |
| `hover <selector>` | 悬停元素 |
| `fill <selector> <value>` | 填充输入框 |
| `type <selector> <text>` | 逐字输入文本 |
| `select <selector> <value>` | 选择下拉选项 |
| `right-click <selector>` | 右键点击（上下文菜单） |
| `dblclick <selector>` | 双击元素 |
| `keyboard <key>` | 按键/组合键（如 `Ctrl+A`、`Enter`、`Tab`） |
| `upload <selector> <file>` | 上传文件 |
| `eval <script>` | 执行 JavaScript |

### 提取

| 命令 | 说明 |
|------|------|
| `text` | 提取页面文本 |
| `screenshot [path]` | 截图 |
| `elements <selector>` | 查找元素 |
| `pdf [file]` | 保存为 PDF（仅 Chromium，标志：`--landscape`、`--format`） |

### 工具

| 命令 | 说明 |
|------|------|
| `wait <selector>` | 等待元素出现 |
| `scroll <up\|down>` | 滚动页面 |
| `pick <x> <y> [--depth=N]` | 检查坐标处元素，返回 DOM 层级 |

### 标签页

| 命令 | 说明 |
|------|------|
| `tab-new` | 新建标签页 |
| `tab-list` | 列出标签页 |
| `tab-switch <id>` | 切换标签页 |
| `tab-close [id]` | 关闭标签页 |

### 对话框

| 命令 | 说明 |
|------|------|
| `dialog-status` | 检查待处理对话框 |
| `dialog-accept [value]` | 接受对话框 |
| `dialog-dismiss` | 关闭对话框 |

### Cookie 和会话

| 命令 | 说明 |
|------|------|
| `cookie list` | 列出已保存的 Cookie |
| `cookie clear [domain]` | 清除指定域名的 Cookie |
| `cookie clear --all` | 清除所有 Cookie |
| `session-list` | 列出活跃会话 |
| `stop` | 停止服务器（Cookie 自动保存） |

## 选择器语法

| 类型 | 示例 | 说明 |
|------|------|------|
| CSS | `#username`, `input[name=email]` | 标准 CSS 选择器 |
| 文本 | `text=Submit` | 包含指定文本的元素 |
| 角色 | `role=button` | ARIA 角色选择器 |
| XPath | `xpath=//div[@id="main"]` | XPath 表达式 |

## 多会话支持

运行隔离的浏览器会话，实现并行 Agent 执行：

```bash
# 每个会话拥有独立的浏览器实例
browser-cli --session agent-1 navigate https://site1.com
browser-cli --session agent-2 navigate https://site2.com

# 列出所有会话
browser-cli session-list

# 停止指定会话
browser-cli --session agent-1 stop
```

## Web Component 支持

### `smart-click` — 智能点击 Web Components

Web Components 通常使用内部回调而非标准 DOM 事件。`smart-click` 自动检测并调用：

```bash
browser-cli smart-click "custom-button"
browser-cli smart-click "[data-action=publish]"
```

检测模式：`_on*`、`_handle*`、`handle*`、`_click`、`_submit`、`_action`

### `pick` — 探索元素内部结构

检查指定坐标处的元素，发现其内部结构：

```bash
browser-cli pick 500 300 --depth=5
```

返回标签名、选择器、检测到的方法、Shadow DOM 结构和建议。

## 对话框检测

Browser-CLI 检测 JavaScript 对话框，由 AI 决定如何处理：

```bash
browser-cli dialog-status
# → {"dialog": {"type": "confirm", "message": "Are you sure?"}}

browser-cli dialog-accept     # 接受
browser-cli dialog-dismiss    # 关闭
```

支持类型：`alert`、`confirm`、`prompt`、`beforeunload`

> 注意：自定义 HTML 弹窗不会被检测 — 请使用元素选择器处理。

## 输出格式

### JSON（推荐 AI 使用）

```json
{"command": "navigate", "status": "success", "data": {"url": "https://example.com/", "title": "Example Domain"}}
```

### Markdown（人类可读）

```
## Navigate
- Status: success
- URL: https://example.com/
- Title: Example Domain
```

## AI 集成指南

### 内置集成文件

| 文件 | 工具 | 说明 |
|------|------|------|
| `skills/browser-cli/SKILL.md` | GAL | 完整 skill 定义和使用模式 |
| `.claude/commands/browser.md` | Claude Code | 自定义斜杠命令 |
| `.codex/skills/browser-cli.md` | OpenAI Codex | Skill 参考文档 |
| `AGENTS.md` | Cursor, Windsurf 等 | Agent 指令文件 |

### AI Agent 最佳实践

1. **服务器自动启动** — 直接运行命令，无需手动管理
2. **使用 `--output json`** — 结构化输出，方便可靠解析
3. **检查 `status` 字段** — 始终为 `"success"` 或 `"error"`
4. **处理对话框** — 点击后检查 `dialog-status`
5. **使用 `--session`** — 隔离并行 Agent 任务
6. **完成后调用 `stop`** — 释放浏览器资源
7. **使用 `--proxy`** — 在防火墙后需显式传入代理 URL
8. **使用 `smart-click`** — 普通 `click` 对自定义元素无效时使用

### 示例：AI Agent 工作流

```bash
# 1. 导航并登录
browser-cli navigate https://login.example.com
browser-cli fill "#username" "user"
browser-cli fill "#password" "pass"
browser-cli click "button[type=submit]"

# 2. 等待页面加载
browser-cli wait ".dashboard"

# 3. 提取数据
browser-cli text
browser-cli eval "JSON.stringify(Array.from(document.querySelectorAll('.item')).map(e=>e.textContent))"

# 4. 导出
browser-cli screenshot result.png
browser-cli pdf report.pdf

# 5. 清理
browser-cli stop
```

## 存储路径

| 资源 | 路径 |
|------|------|
| Cookie | `/tmp/browser-cli/cookies/<domain>.json` |
| 会话 | `/tmp/browser-cli/sessions/<session-id>/server.sock` |

## 系统要求

- Go 1.21+
- Playwright 浏览器（`make setup-browsers`）

## 许可证

[Apache License 2.0](LICENSE)

Copyright 2024 zmysysz
