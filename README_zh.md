# Browser-CLI

[English](README.md) | [中文](README_zh.md)

一个强大的浏览器自动化命令行工具，专为 AI Agent 和自动化工作流设计。

## 为什么选择 Browser-CLI？

- **AI 优先设计** - 结构化 JSON 输出，清晰的命令语义，完美适配 LLM 集成
- **自动管理服务器** - 浏览器服务器按需自动启动，无需手动管理
- **多 Agent 支持** - 隔离会话支持并行 Agent 执行
- **Cookie 持久化** - 自动保存/加载 Cookie，维持登录状态
- **对话框检测** - 检测并处理 JavaScript 对话框，支持 AI 决策
- **跨浏览器** - 通过 Playwright 支持 Chromium、Firefox、WebKit

## 安装

```bash
# 克隆并构建
git clone https://github.com/zmysysz/browser-cli
cd browser-cli
make build

# 安装 Playwright 浏览器（仅首次需要）
make setup-browsers

# 安装到系统
make install
```

## 快速开始

### 简单命令（自动服务器管理）

浏览器服务器会在你运行任何命令时自动启动：

```bash
# 导航 - 服务器自动启动
browser-cli navigate https://example.com

# 截图 - 使用已有的服务器
browser-cli screenshot /tmp/page.png

# 提取文本
browser-cli text

# 完成后停止服务器
browser-cli stop
```

### Run 命令（多步操作）

在单条命令中执行多个操作：

```bash
browser-cli run "navigate https://example.com; click a; text"
```

## 所有命令

### 全局选项

```
--browser, -b     浏览器类型：chromium、firefox、webkit（默认：chromium）
--headless        无头模式（默认：true）
--timeout, -t     超时时间（默认：30s）
--output, -o      输出格式：json、markdown（默认：markdown）
--session, -s     隔离浏览器实例的会话 ID（必填）
--proxy           代理服务器 URL（如 http://proxy.example.com:8080 或 socks5://proxy:1080）
```

### 导航命令

| 命令 | 说明 |
|------|------|
| `navigate <url>` | 导航到 URL |
| `back` | 后退 |
| `forward` | 前进 |
| `reload` | 刷新页面 |

### 交互命令

| 命令 | 说明 |
|------|------|
| `click <selector>` | 点击元素 |
| `click-js <selector>` | 使用 JavaScript 点击（绕过可见性检查） |
| `smart-click <selector>` | 智能点击 Web Components（自动检测内部方法） |
| `hover <selector>` | 悬停元素（显示虚拟光标） |
| `fill <selector> <value>` | 填充输入框 |
| `type <selector> <text>` | 逐字输入文本 |
| `select <selector> <value>` | 选择下拉选项 |
| `eval <script>` | 执行 JavaScript |
| `pick <x> <y> [--depth=N]` | 拾取坐标处元素，返回 DOM 层级和方法 |
| `right-click <selector>` | 右键点击元素（上下文菜单） |
| `dblclick <selector>` | 双击元素 |
| `upload <selector> <file>` | 上传文件到文件输入框 |
| `keyboard <key>` | 按下键盘按键/组合键（如 Ctrl+A、Enter） |

### Web Component 支持

Browser-CLI 提供了专门处理 Web Components（自定义元素）的特殊命令：

#### `smart-click` - 自动检测并点击 Web Components

Web Components 通常使用内部回调函数而非标准 DOM 事件。`smart-click` 会自动检测并调用这些方法：

```bash
# 适用于 <custom-button>、<xhs-publish-btn> 等自定义元素
browser-cli smart-click "custom-button"
browser-cli smart-click "[data-action=publish]"
```

检测模式：`_on*`、`_handle*`、`handle*`、`_click`、`_submit`、`_action`

#### `pick` - 探索元素内部结构

使用 `pick` 检查指定坐标处的元素，发现其内部结构：

```bash
# 拾取指定坐标处的元素（从截图中获取坐标）
browser-cli pick 500 300 --depth=5

# 返回结果：
{
  "target": {
    "tagName": "CUSTOM-BUTTON",
    "selector": "custom-button",
    "methods": ["_onClick", "_onPublish"],  // 检测到可调用方法！
    "attributes": {"data-action": "publish"}
  },
  "ancestors": [
    {"level": 1, "tagName": "DIV", "selector": ".toolbar", "children": ["custom-button", "save-btn"]}
  ],
  "shadowDOM": {"host": "custom-button", "children": ["button.internal"]},
  "suggestions": ["Web Component detected: try smart-click"]
}
```

使用场景：
- 发现 Web Components 上的内部方法，如 `_onPublish`
- 查找嵌套元素的正确选择器
- 了解 Shadow DOM 结构
- 调试 `click()` 在自定义元素上不起作用的原因

### 提取命令

| 命令 | 说明 |
|------|------|
| `screenshot [path]` | 截图 |
| `text` | 提取页面文本 |
| `elements <selector>` | 查找元素 |
| `pdf [file]` | 保存为 PDF（仅 Chromium，标志：--landscape, --format） |

### 文件上传

使用 `upload` 命令将文件上传到文件输入框：

```bash
# 上传单个文件
browser-cli upload "#file-input" ./document.pdf

# 上传到指定类型的输入框
browser-cli upload "input[type=file]" /tmp/image.png
```

### PDF 导出

使用 `pdf` 命令将当前页面保存为 PDF 文件（仅支持 Chromium 浏览器）：

```bash
# 默认设置保存 PDF（A4，纵向）
browser-cli pdf

# 自定义输出路径
browser-cli pdf report.pdf

# 横向和 Letter 格式
browser-cli pdf --landscape --format Letter page.pdf
```

### 键盘快捷键

使用 `keyboard` 命令模拟键盘按键和组合键操作：

```bash
# 单键
browser-cli keyboard "Enter"
browser-cli keyboard "Escape"
browser-cli keyboard "Tab"

# 组合键
browser-cli keyboard "Ctrl+A"        # 全选
browser-cli keyboard "Ctrl+C"        # 复制
browser-cli keyboard "Ctrl+V"        # 粘贴
browser-cli keyboard "Ctrl+Shift+I"  # 开发者工具
```

### 高级点击

除了常规点击外，Browser-CLI 还支持右键点击和双击操作：

```bash
# 右键点击打开上下文菜单
browser-cli right-click "#menu-item"

# 双击选择或激活
browser-cli dblclick "#item"
```

### 工具命令

| 命令 | 说明 |
|------|------|
| `wait <selector>` | 等待元素出现 |
| `scroll <direction>` | 滚动页面（up/down） |

### 标签页命令

| 命令 | 说明 |
|------|------|
| `tab-new` | 创建新标签页 |
| `tab-switch <id>` | 切换到指定标签页 |
| `tab-list` | 列出所有标签页 |
| `tab-close [id]` | 关闭标签页 |

### 对话框命令

| 命令 | 说明 |
|------|------|
| `dialog-status` | 检查待处理对话框 |
| `dialog-accept [value]` | 接受对话框 |
| `dialog-dismiss` | 关闭对话框 |

### 服务器命令

| 命令 | 说明 |
|------|------|
| `server` | 手动启动服务器（前台运行） |
| `status` | 检查服务器状态 |
| `stop` | 停止服务器并保存 Cookie |
| `session-list` | 列出所有活跃会话 |

### Cookie 命令

| 命令 | 说明 |
|------|------|
| `cookie list` | 列出已保存的 Cookie |
| `cookie clear [domain]` | 清除指定域名的 Cookie |
| `cookie clear --all` | 清除所有 Cookie |

## 多会话支持

运行多个隔离的浏览器会话，实现并行 Agent 执行：

```bash
# 每个会话拥有独立的浏览器实例
browser-cli --session agent-1 navigate https://example.com
browser-cli --session agent-2 navigate https://google.com

# 列出所有会话
browser-cli session-list

# 停止指定会话
browser-cli --session agent-1 stop
```

### 会话隔离

每个会话拥有：
- 独立的浏览器实例
- 独立的标签页管理
- 共享的 Cookie 存储（登录状态保持）

## Cookie 管理

Cookie 会自动保存和加载，在会话之间维持登录状态：

```bash
# 查看已保存的 Cookie
browser-cli cookie list

# 清除指定域名的 Cookie
browser-cli cookie clear example.com

# 清除所有 Cookie
browser-cli cookie clear --all
```

### Cookie 存储

Cookie 存储路径为 `/tmp/browser-cli/cookies/<domain>.json`

## 对话框检测

Browser-CLI 能够检测 JavaScript 对话框，并让 AI 决定如何处理：

```bash
# 检查待处理的对话框
browser-cli dialog-status
# 返回：{"dialog": {"type": "confirm", "message": "Are you sure?"}}

# 接受对话框
browser-cli dialog-accept

# 关闭对话框
browser-cli dialog-dismiss

# 接受提示框并输入值
browser-cli dialog-accept "user input"
```

### 支持的对话框类型

| 类型 | 说明 |
|------|------|
| `alert` | 简单提示框，只能接受 |
| `confirm` | 确认对话框，可接受或关闭 |
| `prompt` | 输入对话框，可接受并传入值 |
| `beforeunload` | 页面离开确认 |

注意：自定义 HTML 弹窗（如隐私政策对话框）不会被检测到。请使用元素选择器来处理它们。

## 输出格式

### JSON（推荐 AI 使用）

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

### Markdown（人类可读）

```
## Navigate
- Status: success
- URL: https://example.com/
- Title: Example Domain
```

## 选择器语法

Browser-CLI 支持多种选择器格式：

| 选择器 | 示例 | 说明 |
|--------|------|------|
| CSS | `#username` | CSS 选择器 |
| 文本 | `text=Submit` | 包含指定文本的元素 |
| 角色 | `role=button` | 按 ARIA 角色选择元素 |
| XPath | `xpath=//div[@id="main"]` | XPath 选择器 |

## AI 集成最佳实践

1. **自动服务器** - 服务器自动启动，直接运行命令即可
2. **使用 JSON 输出** - 通过 `--output json` 以编程方式解析结果
3. **检查状态字段** - 每次操作返回 "success" 或 "error"
4. **处理对话框** - 在继续操作前检查 `dialog-status`
5. **使用会话** - 通过 `--session` 隔离并行 Agent 任务
6. **保持登录** - Cookie 自动保存，登录状态持久化

### AI 工作流示例

```bash
# 1. 导航并登录（服务器自动启动）
browser-cli navigate https://login.example.com
browser-cli fill "#username" "user"
browser-cli fill "#password" "pass"
browser-cli click "button[type=submit]"

# 2. 等待登录完成
browser-cli wait ".dashboard"

# 3. 检查对话框
browser-cli dialog-status

# 4. 使用键盘快捷键操作
browser-cli keyboard "Ctrl+A"        # 全选内容
browser-cli keyboard "Ctrl+C"        # 复制到剪贴板
browser-cli keyboard "Enter"         # 确认操作

# 5. 提取数据
browser-cli text
browser-cli screenshot result.png

# 6. 导出为 PDF
browser-cli pdf report.pdf

# 7. 停止服务器（Cookie 自动保存）
browser-cli stop
```

## 存储路径

| 资源 | 路径 |
|------|------|
| Cookie | `/tmp/browser-cli/cookies/<domain>.json` |
| 会话 | `/tmp/browser-cli/sessions/<session-id>/server.sock` |

## 系统要求

- Go 1.21+
- Playwright 浏览器（通过 `make setup-browsers` 安装）

## 许可证

[Apache License 2.0](LICENSE)

Copyright 2024 zmysysz
