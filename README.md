# KineticGo

基于 [Wails](https://wails.io/) 框架开发的本地个人自动化与系统监控平台。支持任务卡片化管理、定时调度、实时日志推送。

![Platform](https://img.shields.io/badge/Platform-macOS%20|%20Windows-blue)
![Go](https://img.shields.io/badge/Go-1.25-green)
![Wails](https://img.shields.io/badge/Wails-v2.12-red)

## 功能特性

### 系统监控

| 任务类型 | 说明 |
|---------|------|
| `system-local_cpu` | 实时 CPU 使用率监控 |
| `system-local_memory` | 实时内存使用率监控 |
| `system-local_network` | 网络活动监控 |
| `system-active_tasks` | 活跃任务数监控 |

### 业务自动化
| 任务类型 | 说明 |
|---------|------|
| `campus_auth` | 校园网自动登录（四川轻化工大学） |
| `652_signin` | 652 签到 |

### 核心能力
- **插件式架构**: 通过 `TaskFactory` 工厂模式实现任务注册与隔离
- **定时调度**: 内置 cron 调度器，支持标准 cron 表达式
- **实时日志**: 通过 Wails Events 实时推送任务日志到前端
- **热启动**: 系统任务开机自动启动，常驻后台
- **执行历史**: 完整的任务执行记录与日志审计

## 技术栈

```
后端    │  Go 1.25 + GORM + SQLite
前端    │  Vue 3 + Vite + Pinia
框架    │  Wails v2
调度    │  robfig/cron/v3
系统信息 │  shirou/gopsutil/v3
```

## 项目结构

```
KineticGo/
├── main.go                 # 应用入口
├── embed.go                # 前端资源嵌入
├── wails.json              # Wails 配置
│
├── internal/
│   ├── model/              # 数据模型
│   │   ├── task.go         # Task, TaskSchedule, TaskExecution, TaskLog
│   │   └── option.go       # TempleConfig 类型定义
│   │
│   ├── repository/         # 数据访问层
│   │   ├── db.go           # GORM 初始化 & AutoMigrate
│   │   └── task_repository.go
│   │
│   ├── service/            # 业务逻辑层
│   │   ├── taskmanage.go    # 任务调度中心（注册、启动、停止）
│   │   ├── scheduler.go     # Cron 调度器封装
│   │   ├── system.go        # 系统监控任务（CPU/内存/网络）
│   │   ├── suse_wifi.go     # 校园网登录
│   │   ├── signin.go        # 652 签到服务
│   │   ├── task_log.go      # 日志推送服务
│   │   └── updater.go       # 自动更新
│   │
│   ├── ocr/                # OCR 识别（ONNX Runtime）
│   │   └── onnx_engine.go
│   │
│   └── wailsapp/           # Wails 应用绑定层
│       ├── app.go          # App 结构体，暴露方法给前端
│       └── wire.go         # Google Wire 依赖注入
│
├── frontend/               # Vue 前端
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   └── components/
│   │       ├── DashboardView.vue    # 首页仪表盘
│   │       ├── TaskCard.vue         # 任务卡片
│   │       ├── TaskConfigForm.vue   # 任务配置表单
│   │       ├── TaskLogsView.vue     # 实时日志
│   │       ├── LogHistoryView.vue   # 历史审计
│   │       ├── LoadTesterView.vue   # 压测工具
│   │       ├── SettingsView.vue     # 设置页
│   │       └── charts/
│   │           └── QpsChart.vue     # QPS 图表
│   └── package.json
│
└── pkg/
    ├── goload/             # 压测引擎
    └── location/           # 位置服务
```

## 核心接口

所有任务必须实现 `TaskInstance` 接口：

```go
type TaskInstance interface {
    Run(ctx context.Context, scheduleId uint) error
    Stop(scheduleId uint) error
}
```

任务通过工厂模式注册：

```go
// 普通任务
taskManage.Register("campus_auth", service.NewSuseWifiService)

// 系统任务（开机自启、常驻后台）
taskManage.RegisterSystem("local_cpu", service.NewCpuService)
```

## 数据模型

### Task（任务模版）
定义平台支持的所有任务类型，作为注册中心。

| 字段 | 类型 | 说明 |
|-----|------|------|
| `ID` | uint | 主键 |
| `Name` | string | 任务名称 |
| `Type` | string | 任务标识（如 `load_test`） |
| `Config` | TempleConfig | 前端表单定义的 JSON |

### TaskSchedule（任务实例）
用户在首页创建的每一个卡片。

| 字段 | 类型 | 说明 |
|-----|------|------|
| `ID` | uint | 主键 |
| `Name` | string | 实例名称 |
| `TaskType` | string | 任务类型（如 `system-local_cpu`） |
| `CronExpr` | string | Cron 表达式 |
| `IsEnabled` | bool | 启用开关 |
| `NextRunTime` | time.Time | 下次执行时间 |

### TaskExecution & TaskLog
记录每次任务触发起止时间、状态及详细流水。

## 前端交互

1. **首页仪表盘**: 顶部展示系统核心指标，主体为任务卡片 Grid
2. **任务卡片**: 统一深色卡片设计，集成启动/停止按钮，内嵌微型控制台实时滚动日志
3. **动态表单**: 根据后端 `TemplateSchema` 动态渲染输入组件
4. **历史审计**: 展示执行历史，点击可查看完整结构化日志

## 构建

### 前置依赖

```bash
# 安装 Node.js 依赖
cd frontend && npm install

# 安装 Go 依赖
go mod download
```

### 开发模式

```bash
# 前端开发服务器
cd frontend && npm run dev

# Wails 开发模式（在项目根目录）
wails dev
```

### 生产构建

```bash
# 编译 macOS
wails build -platform darwin/universal

# 编译 Windows
wails build -platform windows

# 编译 Linux
wails build -platform linux
```

产物位于 `build/` 目录。

## 配置

数据库文件位于应用同目录 `kineticgo.db`，首次启动时 GORM 会自动执行 AutoMigrate。

## License

MIT