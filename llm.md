# 飞牛 fnOS 原生应用开发 - 架构设计与技术决议

## 项目背景
飞牛 fnOS 提供了官方远程访问通道（FN Connect），通过专用域名可直接访问处于内网的 NAS 系统及官方应用。
针对用户部署的自建服务无法直接穿透的问题，本项目利用飞牛开发者文档与工具链，构建原生 Native 飞牛应用，直接接入飞牛**统一网关（Unified Gateway）**，实现无需公网端口映射、无需额外隧道工具、无缝复用 FN Connect 域名的远程终端控制台。

---

## 核心应用：飞牛原生系统终端（Terminal）

### 1. 架构拓扑与数据流转
- **访问入口**：飞牛统一网关将 `/app/terminal` 前缀代理转发至应用本地 Unix Domain Socket：`/var/apps/terminal/target/app.sock`。
- **协议穿透**：HTTP 静态页面与 WebSocket（WSS）长连接复用同一个 Socket，网关自动校验 NAS 登录态并注入用户信息 Header。
- **底层驱动**：基于 Linux 原生 PTY（伪终端）直接调度终端会话，完全脱离 NAS 系统 SSH 守护进程（sshd）依赖，即使 SSH 服务关闭或崩溃仍可作为应急控制台。

```
[浏览器 / FN Connect 远程访问]
       │ (HTTPS / WSS)
       ▼
[飞牛统一网关 (Unified Gateway)]
       │ (校验登录态，注入 X-Trim-Username)
       ▼ (Unix Domain Socket: app.sock)
[Go 终端后台服务 (单文件 ELF)]
       │ (runuser -u <username> -- /bin/bash -l)
       ▼ (Linux PTY: creack/pty)
[宿主机交互式 Shell 会话 (当前登录用户身份)]
```

### 2. 权限与身份隔离模型决议
- **服务常驻身份**：`config/privilege` 声明 `"run-as": "root"`。
  - **核心原因**：守护进程在开机/应用启动时常驻运行，此时尚无具体用户访问。只有 root 进程具备在收到连接时无密码切换子进程 UID（`runuser -u <user>`）的能力；非 root 进程无法切换其他普通用户身份。
- **终端会话身份**：降权为网关当前登录用户。
  - 读取 Header `X-Trim-Username`；
  - 启动终端命令：`runuser -l <username> -c "exec /bin/bash -l"`，进程工作目录设为用户 `$HOME`；
  - 终端初始身份直接为当前登录用户（如 `nas01`，家目录为 `/home/nas01`，UID 为 1000），绝不默认泄露 root 权限；
  - 用户能否执行特权操作，完全取决于其自身在 `/etc/sudoers` 中的配置，需自行输入密码提权。

### 3. 工程命名与结构规范
废弃通用的 `backend` / `frontend` 命名，贴合飞牛工程规范：
- `ui/`：前端 Web 工程（Vue 3 Options API + TypeScript + Vite + xterm.js + pnpm）；
- `pty/`：后端核心驱动工程（Go 1.26 + PTY 管理 + Unix Socket + WebSocket 桥接 + 静态资源 embed）；
- `package/`：飞牛应用打包规范目录（manifest, config, cmd, ui, icons）；
- `scripts/`：Windows / Linux 构建辅助工具（`write_cmd_main.go`、`set_version.go`、`set_platform.go`、`gen_icons.go`）；
- `build.bat`：项目根目录一键多架构打包脚本，支持命令行传入目标版本号（例如 `build.bat 1.0.1`），自动更新 `package/manifest` 与 `ui/package.json`，交叉编译 `amd64` 与 `arm64` 两个平台的二进制，并在 `output/` 目录统一输出规范命名的安装包（如 `output/terminal-1.0.9-amd64.fpk` 与 `output/terminal-1.0.9-arm64.fpk`），根目录不残留无架构后缀或临时文件。


### 4. 退出生命周期与 UTMP 机制决议
- **会话退出捕获（EOF Hook）与操作优化**：
  - 当用户在终端中按下 `Ctrl+D` 或键入 `exit`，bash 进程退出，PTY 读取返回 EOF；
  - Go 后端感知后立即向 WebSocket 发送 `{"type":"exit","code":0}` 控制帧；
  - 前端捕获后将终端标记为断开，并在 xterm 内部输出 `[会话已结束] 终端进程已退出 (按回车键重新连接)`；
  - 界面底部弹出悬浮操作栏提供“重新连接”；支持用户直接敲击回车（Enter）一键重连启动新会话；
  - 移除无用的“关闭窗口”按钮：Web 终端在桌面 iframe 或非 `window.open` 打开的独立标签页中，浏览器安全机制禁止脚本通过 `window.close()` 自行关闭宿主窗口，避免给用户造成误导。

### 5. 移动端与触控交互适配决议
- **视口与防缩放控制**：
  - HTML Viewport 配置 `user-scalable=no, viewport-fit=cover`，彻底消除移动端快速敲击或双击时的缩放抖动，适配 iPhone 底部安全区（`safe-area-inset`）。
- **移动虚拟辅助按键行（Mobile Virtual Bar）**：
  - 手机软键盘缺少终端核心控制键（ESC、TAB、CTRL、ALT、光标方向键、常用特殊符号 `/`、`|`、`-`、`~` 等）；
  - 前端提供吸附于底部的虚拟按键条，并在移动端自适应显示（也可手动展开/收起）；
  - 按键采用 `@touchstart.prevent` 绑定，严格杜绝触控点击导致的软键盘意外失焦收回；
  - 支持**黏性修饰键（Sticky Keys）**：单按 CTRL 或 ALT 激活待命状态，按下下一个字符时自动合成控制字符（如 `CTRL` + `c` 发送 `\x03`）并自动恢复。
- **全局触控聚焦流转**：
  - 无论点击终端黑屏任意区域，自动聚焦底层 xterm textarea，确保手机虚拟键盘随点随起。
- **会话结束「胶囊浮岛」（Capsule Island）交互设计**：
  - 彻底杜绝中文字符在小屏上的逐字竖向换行崩溃：严格声明 `white-space: nowrap` 与 `flex-shrink: 0`，排版严整单行；
  - 响应式文案：移动窄屏（`<= 600px`）精简显示 `🟡 会话已结束` + `[ ↺ 重新连接 ]`，总宽仅 190px、高度 34px，精致小巧；宽屏自动补充 `（按 Enter 快速重连）`；
  - 超椭圆圆角（`border-radius: 9999px`）搭配深色磨砂毛玻璃（`backdrop-filter: blur(16px)`）、微光外边框与金黄色状态呼吸脉冲灯；
  - 全域重连触控热区：手机触屏点击胶囊任意位置或按钮均可触发极速重新连接。

### 6. 消除 Bash 作业控制警告（Job Control 修复）
- **根因**：原先使用 `runuser -l <user> -c "exec /bin/bash -l"`，通过 `-c` 启动子命令导致 Bash 处于非交互子 Shell 模式，无法通过 `ioctl(TIOCSPGRP)` 申请终端前台进程组控制权，引发：
  `bash: cannot set terminal process group (-1): Inappropriate ioctl for device` 与 `bash: no job control in this shell`。
- **优化**：调整为 `runuser -l <user> -s /bin/bash`（直接由 runuser 以登录模式将 PTY 控制台交予 `/bin/bash`），彻底清除上述报错并恢复完整的作业控制功能（支持 `Ctrl+Z` 挂起与 `bg`/`fg` 等作业调度）。

### 7. 关于 `w` / `who` 命令看不到会话的原因
- Linux `w`、`who` 是读取 `/var/run/utmp` 数据库来展现登录用户的；
- 标准由 `sshd` 或 `/bin/login` 写入，而 `runuser` 与现代容器/Web 终端驱动（如 `docker exec`、VS Code Terminal、code-server）默认均绕过 UTMP 写入；
- 这避免了 Web 窗口直接关闭时因无法触发标准 logout 而导致 `/var/run/utmp` 残留僵尸在线会话记录的问题。

### 8. 应用命名规范与大小写敏感决议（Package ID vs Display Name）
- **包名（Package ID / `appname`）**：
  - 严格保持全小写 `terminal`；
  - 绑定系统文件系统安装路径 `/var/apps/terminal/`、网关路由 `/app/terminal`、生命周期脚本环境与 Nginx 规则；
  - 避免因大小写不一致导致的 Nginx 404 Not Found；
- **显示名称（Display Name / `display_name` / `title`）**：
  - 配置为中文名称 `终端`（`manifest` 中 `display_name=终端`，`app/ui/config` 中 `"title": "终端"`）；
  - 飞牛桌面应用图标、应用中心应用卡片与弹出窗口标题均统一以 `终端` 展现给用户，界面亲和且贴合 NAS 系统原生风格。

### 9. 移动端与 iOS WebKit 深度适配决议
- **消除 iOS 原生键盘 Form Assistant（`^` `v` `✓`）栏**：
  - **现象根因**：iOS Safari 在页面存在两个及以上表单元素（`<input>` / `<textarea>` / `<select>`）时，会自动在虚拟键盘上方注入系统级原生 `UIInputAccessoryView`（带有上下表单切换箭头 `^` `v` 与完成按钮 `✓`）；由于工具栏原本使用了原生 `<select>`，按 `^` 会把焦点转移到 `<select>` 上并唤起 iOS 系统级滚轮选择器。该控件属于 iOS 操作系统原生层，在浏览器 DOM 树中不存在；
  - **解决方案**：将原生 `<select>` 替换为纯非表单的自定义下拉组件（`<button>` + 悬浮 `<div>`），使页面上只存在 xterm.js 的单个隐藏输入 capture 节点。iOS WebKit 检测到页面无多表单控件后，彻底消除 `^` `v` 切换栏。
- **iOS 虚拟键盘升起时辅助键栏遮挡问题（Visual Viewport 驱动）**：
  - **现象根因**：iOS Safari 在软键盘升起时，**不会缩小布局视口（Layout Viewport）**，`100vh` 依然等于屏幕全高；因此底部绝对/弹性定位的辅助键条会被 iOS 键盘硬生生盖在下方；
  - **解决方案**：通过 `window.visualViewport` 监听动态视口尺寸。当键盘升起时，动态将应用根容器高度收缩至 `visualViewport.height`，使辅助键栏精准浮贴在软键盘正上方，并触发 `xterm.fit()` 自适应终端行数。
- **软键盘收起后页面上移与底部间隙消失问题**：
  - **现象根因**：移动端在聚焦输入框时可能会自动轻微向上滚屏（`window.scrollY > 0`），而软键盘收起时 WebKit 不会自动滚回顶部，导致页面底部被推移至屏幕外，底部 Home Indicator 区域的边距被吞没；且先前辅助栏使用固定高度 `44px` 挤压了 `env(safe-area-inset-bottom)`；
  - **解决方案**：
    1. 在 `html, body` 声明 `position: fixed; inset: 0; overflow: hidden;`，禁止整个视口的弹性滚动；
    2. 在视口缩放与失焦事件中显式执行 `window.scrollTo(0, 0)` 复位；
    3. 辅助键栏高度采用动态计算：键盘未弹出时高度为 `calc(44px + env(safe-area-inset-bottom, 0px))` 且保持 `padding-bottom: max(6px, env(safe-area-inset-bottom))`，为 iPhone 底部小横条留出充裕优雅的呼吸间距；键盘弹出时动态切换至紧凑模式（`height: 42px; padding-bottom: 4px;`），紧贴键盘上方。
