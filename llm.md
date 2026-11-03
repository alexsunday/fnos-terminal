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
- `scripts/`：Windows / Linux 构建辅助工具（`write_cmd_main.go`、`set_version.go`、`gen_icons.go`）；
- `build.bat` / `build.sh`：项目根目录一键打包脚本，支持命令行传入目标版本号（例如 `build.bat 1.0.1` 或 `./build.sh 1.0.1`），自动同步更新 `package/manifest` 与 `ui/package.json`，并输出带版本命名的安装包产物（如 `terminal-1.0.1.fpk`）。


### 4. 退出生命周期与 UTMP 机制决议
- **会话退出捕获（EOF Hook）**：
  - 当用户在终端中按下 `Ctrl+D` 或键入 `exit`，bash 进程退出，PTY 读取返回 EOF；
  - Go 后端感知后立即向 WebSocket 发送 `{"type":"exit","code":0}` 控制帧；
  - 前端捕获后将终端标记为断开，并在 xterm 内部输出 `[会话已结束] 终端进程已退出 (按回车键重新连接)`；
  - 界面底部弹出悬浮操作栏，提供快捷“重新连接”与“关闭窗口”操作；支持用户键盘敲击回车直接一键重连启动新会话。
- **关于 `w` / `who` 命令看不到会话的原因**：
  - Linux `w`、`who` 是读取 `/var/run/utmp` 数据库来展现登录用户的；
  - 标准由 `sshd` 或 `/bin/login` 写入，而 `runuser` 与现代容器/Web 终端驱动（如 `docker exec`、VS Code Terminal、code-server）默认均绕过 UTMP 写入；
  - 这避免了 Web 窗口直接关闭时因无法触发标准 logout 而导致 `/var/run/utmp` 残留僵尸在线会话记录的问题。
