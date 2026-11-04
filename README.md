# 飞牛 fnOS 原生终端应用（Terminal）

<p align="center">
  <img src="./package/icons/terminal-256.png" alt="飞牛终端 Logo" width="128" height="128" />
</p>

<p align="center">
  <b>专为飞牛私有云（fnOS）打造的现代化原生网页终端工具</b><br>
  无需开放 SSH 端口 · 深度融合 FN Connect 远程穿透 · 严格即时降权安全模型 · 移动端全触控体验
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-fnOS_%3E%3D_1.1.3100-blue.svg" alt="fnOS" />
  <img src="https://img.shields.io/badge/Arch-amd64_%7C_arm64-success.svg" alt="Architecture" />
  <img src="https://img.shields.io/badge/License-AGPL--3.0-blue.svg" alt="License" />
</p>

## 📖 项目简介

**「飞牛终端」** 是一款专为飞牛私有云（fnOS）原生设计的交互式网页终端应用（FPK）。

传统 NAS 在外网访问终端通常需要“开启 SSH 服务 + 路由器端口转发/公网 IP 映射”，这不仅配置繁琐，还会让宿主机面临严重的公网端口扫描与暴力破解威胁。

本应用直接接入飞牛系统的**统一网关（Unified Gateway）**，利用 Linux 原生 PTY（伪终端）技术，在不需要开启系统 SSH 服务的物理前提下，为当前登录用户无缝提供随时随地、开箱即用的 Linux 终端交互能力。

## 🌟 核心特性

- 🔒 **零端口暴露 · 无 SSH 依赖**：
  - 宿主机**无需开放 22 或任何自定义 SSH 端口**，也无需配置路由器端口映射；
  - 即使系统设置中已完全关闭或停止了 SSH 服务，终端应用依然稳定可用，可充当核心应急控制台。
- ⚡ **深度融合网关 · FN Connect 原生畅连**：
  - 深度接入飞牛统一认证机制，已登录飞牛系统的用户直接单点进入终端；
  - 完美适配官方 **FN Connect** 远程穿透与专用域名，异地访问宛如局域网般丝滑。
- 🛡️ **严格安全模型 · 最小特权与即时降权**：
  - **特权仅用于瞬间会话分发**：后台守护进程仅利用特权确认身份；
  - **即时降权执行**：拉起终端时瞬时降权为当前网关登录用户（`runuser` 隔离），实际交互的 Bash 进程完全属于受限普通用户，无法越权访问他人隐私或修改系统关键配置；
  - **零存储索取**：不申请任何用户存储卷（Shared Folders）读写权限，无文件泄露风险。
- 📱 **移动端与触控极致适配**：
  - **移动虚拟辅助按键栏**：内置 `Ctrl`、`Alt`、`Esc`、`Tab`、方向键等移动软键盘缺失的终端核心控制键；
  - **粘性修饰键（Sticky Keys）**：支持单按 `Ctrl` / `Alt` 后输入字符自动组合（如 `Ctrl` + `c` 发送中断信号）；
  - **iOS WebKit 深度优化**：动态监听 `visualViewport`，键盘升起时辅助栏自适应浮贴软键盘上方；消除 iOS 系统多表单切换栏干扰；
  - **全触控唤起**：点击黑屏任意区域自动聚焦软键盘，告别触屏失焦烦恼。
- 📦 **多架构支持**：
  - 官方支持 `amd64` (x86_64) 与 `arm64` 平台，单二进制轻量级部署。

## 🏗️ 架构拓扑与数据流

```text
[ 浏览器客户端 / 手机移动端 ]
       │ (HTTPS / WSS 安全加密长连接)
       ▼
[ 飞牛统一网关 (Unified Gateway / FN Connect) ]
       │ (1. 校验飞牛登录态；2. 注入 X-Trim-Username)
       ▼ (Unix Domain Socket 进程间通信: /var/apps/terminal/target/app.sock)
[ Go 终端守护进程 (单文件原生 ELF) ]
       │ (读取用户名，瞬间执行降权: runuser -l <username> -s /bin/bash)
       ▼ (Linux 原生 PTY: creack/pty)
[ 宿主机受限 Bash 会话 (当前登录用户身份) ]
```

## 🛡️ 安全与合规说明

针对飞牛应用开放平台 [《权限配置与开发最佳实践》（Privilege）](https://developer.fnnas.com/docs/core-concepts/privilege/)，本项目作如下安全设计保证：

| 维度 | 设计方案 | 安全保障 |
| :--- | :--- | :--- |
| **网络边界** | 不依赖 `sshd`，不监听公网 TCP 端口 | 零公网端口暴露面，彻底免疫公网 SSH 爆破与扫描探测 |
| **数据传输** | 复用飞牛官方网关 HTTPS / WSS（TLS）加密隧道 | 传输链路具备完整防嗅探、防中间人劫持能力 |
| **运行权限 (`config/privilege`)** | 后台守护进程声明 `run-as: root` | 因 Linux 权限隔离限制，沙盒用户无权切换至其他登录用户的 UID；root 仅用于初始守护与瞬间切换身份，**已无可更窄方案** |
| **会话权限** | 启动 Bash 瞬间降权为登录用户 (`runuser -l`) | 终端会话以受限普通用户执行，无法越权修改系统关键配置或提权 |
| **存储访问** | 零存储权限申请（未申请任何存储池读写） | Web 端无文件浏览或下载接口，不触碰任何用户存储文件 |

## 📁 目录结构

```text
.
├── pty/              # 后端驱动工程（Go 1.26 + PTY 管理 + Unix Socket + WebSocket 桥接 + 静态资源嵌入）
├── ui/               # 前端 Web 工程（Vue 3 Options API + TypeScript + Vite + xterm.js）
├── package/          # 飞牛应用安装包模板结构（manifest, config, cmd, icons）
├── scripts/          # 打包自动化脚本（版本注入、平台架构切换、图标生成）
├── build.bat         # Windows 一键多架构编译与 FPK 打包脚本
├── output/           # 打包生成的 .fpk 安装包输出目录
├── llm.md            # 项目架构设计细节与技术决议记录
└── README.md         # 项目使用与说明文档
```

## 🛠️ 构建与打包

### 前置要求
1. **Node.js** (v18+) 及 **pnpm**；
2. **Go** (v1.22+) 并支持交叉编译；
3. 飞牛官方打包工具 **`fnpack.exe`**（已内置于工程根目录）。

### 一键打包（推荐）
在工程根目录下执行构建批处理脚本，支持指定版本号（如 `1.0.9`）：

```cmd
build.bat 1.0.9
```

构建脚本将自动执行以下流程：
1. 编译前端生产静态资源（`ui/dist`）；
2. 静态资源自动嵌入 Go 二进制，并交叉编译 `linux/amd64` 与 `linux/arm64` 原生 ELF 执行文件；
3. 更新 `package/manifest` 与版本配置；
4. 调用 `fnpack` 在 `output/` 目录下生成各架构专用的安装包：
   - `output/terminal-1.0.9-amd64.fpk`
   - `output/terminal-1.0.9-arm64.fpk`

## 💻 安装与使用

1. 打开飞牛私有云（fnOS）Web 桌面；
2. 进入 **应用中心** -> **手动安装**；
3. 选择对应硬件架构的 `.fpk` 文件（如 x86 设备选择 `amd64.fpk`，ARM 架构设备选择 `arm64.fpk`）并上传安装；
4. 安装完成后，在桌面点击 **「终端」** 图标即可直接打开使用。

## 📄 开源许可证

本项目基于 [GNU Affero General Public License v3.0 (AGPL-3.0)](./LICENSE) 协议开源。
