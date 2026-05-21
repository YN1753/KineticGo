# KineticGo

**KineticGo** 是一款基于 [Wails v2](https://wails.io/) 框架开发的现代桌面个人自动化与系统运维工具。它采用 **Go + Vue 3** 架构，旨在通过极简的 **Slate/Zinc** 风格界面，为开发者和校园用户提供一键式的高效自动化体验。

![Platform](https://img.shields.io/badge/Platform-Windows%20|%20macOS%20|%20Linux-blue)
![Go](https://img.shields.io/badge/Go-1.25+-green)
![Wails](https://img.shields.io/badge/Wails-v2.12-red)
![Vue](https://img.shields.io/badge/Vue-3.x-brightgreen)

---

## ✨ 核心特性

### 🚀 应用启动舱 (App Launcher)
- **移动端图标风格**: 采用 1:1 紧凑型图标网格，告别凌乱的任务卡片。
- **跨平台异步拉起**: 一键秒开本地 `.exe`、脚本、或网页 URL，不卡死主进程。
- **高度自定义**: 支持可视化选择 7 种图标（火箭、闪电、代码等）与 6 种经过调色的主题色彩。
- **智能标签**: 支持自定义快捷入口名称，方便辨识。

### 🏫 校园服务 (School Services)
- **校园网自动连**: 实时监测网络状态，掉线自动执行登录协议（支持四川轻化工大学）。
- **652 自动签到**: 集成 CAS 统一身份认证，定时自动完成位置签到。
- **专属管理列**: 独立的校园服务管理区域，一键注册，状态实时同步。

### 🛠️ 端口大盘 (Port Killer)
- **全量排查**: 跨平台扫描系统占用端口（Windows 使用 netstat，Unix 使用 lsof）。
- **进程防护**: 智能识别系统核心进程与用户应用，防止误杀。
- **一键释放**: 针对非核心进程支持一键强杀，解决端口冲突无需进入命令行。

### 📊 系统实时看板
- **实时推流**: 采用 Wails Events 驱动，秒级更新 CPU、内存、实时网速（上传/下载）及活跃任务数。
- **会话日志终端**: 内置高性能日志推送服务，带内存节流机制，防止高频日志压垮 UI 渲染。

---

## 🛠️ 技术架构

### 后端 (Go)
- **核心框架**: Wails v2
- **依赖注入**: Google Wire
- **数据库**: SQLite + GORM
- **定时调度**: robfig/cron/v3
- **跨平台兼容**: 封装了 `SysProcAttr` 抽象层，在 Windows 下支持后台静默启动（隐藏 CMD 黑窗口）。

### 前端 (Vue 3)
- **界面风格**: 极简 Slate/Zinc 设计语言
- **组件库**: 原生 Vue 3 + Tailwind CSS
- **图标库**: Lucide-vue-next
- **响应式**: 适配 1280px+ 宽屏展示，三列布局一目了然。

---

## 📂 项目结构

```text
KineticGo/
├── internal/
│   ├── model/           # 数据模型 (Task, Schedule, Config)
│   ├── repository/      # 数据访问层 (GORM 初始化 & 种子数据)
│   ├── service/         # 业务逻辑 (任务调度、应用拉起、端口排查、日志推送)
│   └── wailsapp/        # Wails 绑定与 Wire 注入
├── frontend/
│   ├── src/
│   │   ├── components/  # Dashboard, TaskCard, TaskConfigForm, VisualPickers
│   │   └── composables/ # API 封装
│   └── vite.config.js
├── main.go              # 应用入口 (窗口尺寸配置 1280x800)
├── wails.json           # Wails 编译配置
└── README.md
```

---

## 🚀 快速开始

### 前置要求
- [Go](https://golang.org/dl/) 1.25+
- [Node.js](https://nodejs.org/) 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### 开发环境
```bash
# 安装依赖
go mod download
cd frontend && npm install

# 启动开发服务器
wails dev
```

### 生产构建
```bash
# 构建 Windows 版本
wails build -platform windows

# 构建 macOS 版本
wails build -platform darwin/universal

# 构建 Linux 版本
wails build -platform linux
```

---

## ⚙️ 进阶开发

### 注册新任务类型
所有任务需实现 `TaskInstance` 接口。在 `internal/wailsapp/app.go` 中通过 `Register` 方法进行挂载：

```go
// 注册业务任务
a.taskManage.Register("your_task_type", service.NewYourService)

// 注册系统常驻任务
a.taskManage.RegisterSystem("your_sys_task", service.NewYourSysService)
```

---

## 📄 License
基于 MIT 协议开源。

---
*Powered by Gemini CLI & KineticGo Team*
