# 飞牛终端 (fnOS Terminal)

[![fnOS](https://img.shields.io/badge/Platform-fnOS_%3E%3D_1.1.3100-blue.svg)](https://www.fnnas.com/)
[![Architecture](https://img.shields.io/badge/Arch-amd64_%7C_arm64-success.svg)](#构建与打包)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](./LICENSE)

飞牛 fnOS 的原生网页终端应用（FPK），直接接入系统统一网关与 FN Connect 远程访问，无需在路由器上做端口转发或在系统内开启 SSH 端口。

## 特性

- **免开 SSH 端口**：直接基于 Linux PTY 运行，不依赖宿主机 `sshd` 进程；系统关闭 SSH 服务时依然可用。
- **网关登录集成**：接入飞牛统一网关，直接复用当前登录用户的会话，支持 FN Connect 域名远程访问。
- **会话降权运行**：
  - 守护进程以 root 运行仅用于在收到连接时切换身份；
  - 启动 Bash 时立即通过 `runuser` 降权至当前登录用户，终端内不默认提供 root 权限；
  - 应用不申请任何用户存储卷（Shared Folders）的读写权限。
- **移动端辅助按键**：
  - 提供底部虚拟功能键栏（Ctrl、Alt、Esc、Tab、方向键及常用字符）；
  - 支持 Ctrl / Alt 粘性组合键（按一下激活，再按字符发送组合键）；
  - 监听 `visualViewport` 适配 iOS Safari 等移动端软键盘弹出时的视口收缩。
- **多架构支持**：支持 x86_64 (`amd64`) 和 AArch64 (`arm64`) 架构。

## 运行原理

```text
[ 浏览器 / 移动端 ]
       │ (HTTPS / WSS)
       ▼
[ 飞牛统一网关 (Unified Gateway / FN Connect) ]
       │ (校验登录态，注入 X-Trim-Username)
       ▼ (Unix Domain Socket: /var/apps/terminal/target/app.sock)
[ 终端服务 ]
       │ (runuser -l <username> -s /bin/bash)
       ▼ (Linux PTY: creack/pty)
[ 宿主机交互式 Shell 会话 (当前登录用户身份) ]
```

## 权限设计说明

根据飞牛应用开发平台的权限规范，本应用的权限边界设计如下：

| 项目 | 说明 |
| :--- | :--- |
| **网络端口** | 不对外监听任何 TCP 端口，仅通过本地 Unix Socket 与系统网关通信。 |
| **运行权限 (`run-as`)** | 配置为 `root`。因 Linux 权限限制，普通包用户无法切换为其他任意登录用户的 UID。 |
| **实际终端权限** | 服务端调用 `runuser -l <user>` 启动 Bash，终端进程实际运行在普通用户权限下。 |
| **存储与外设** | 未申请任何存储池读写权限，Web 端也不提供文件读写或下载接口。 |

## 目录结构

```text
.
├── pty/          # 后端核心（Go + PTY 管理 + Unix Socket + WebSocket + 前端静态资源 embed）
├── ui/           # 前端界面（Vue 3 Options API + TypeScript + Vite + xterm.js）
├── package/      # 飞牛应用包模板（manifest, config, cmd, icons）
├── scripts/      # 打包辅助脚本（版本写入、架构切换等）
├── build.bat     # Windows 一键编译与打包批处理
├── output/       # 打包生成的 .fpk 目录
└── llm.md        # 架构设计与技术决议记录
```

## 构建与打包

### 环境依赖
- **Node.js** (>= 18) 与 **pnpm**
- **Go** (>= 1.22)
- **`fnpack.exe`**（已存放在根目录）

### 执行打包
在项目根目录运行 `build.bat` 并传入版本号：

```cmd
build.bat 1.0.9
```

构建脚本会自动完成前端编译、Go 交叉编译（amd64 / arm64）并调用 `fnpack`，在 `output/` 目录输出：
- `output/terminal-1.0.9-amd64.fpk`
- `output/terminal-1.0.9-arm64.fpk`

## 安装方法

1. 登录飞牛 fnOS 桌面；
2. 打开 **应用中心** -> **手动安装**；
3. 上传对应架构的 `.fpk` 文件完成安装；
4. 安装完成后点击桌面上的 **「终端」** 图标即可使用。

## 开源协议

本项目基于 [GNU Affero General Public License v3.0 (AGPL-3.0)](./LICENSE) 协议开源。
