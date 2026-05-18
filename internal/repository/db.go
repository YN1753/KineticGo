package repository

import (
	"kineticgo/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func DbInit(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic("链接数据库失败" + err.Error())
	}
	err = db.AutoMigrate(&model.Task{}, &model.TaskExecution{}, &model.TaskLog{}, &model.TaskSchedule{})
	if err != nil {
		panic("自动迁移表失败" + err.Error())
	}
	err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_system_task 
                   ON task_schedules(task_type) 
                   WHERE task_type LIKE 'system-%'`).Error
	if err != nil {
		panic("创建系统任务唯一索引失败: " + err.Error())
	}
	DbSeedInit(db)
	return db
}
func DbSeedInit(db *gorm.DB) {
	var existingTemplateNames []string
	db.Model(&model.Task{}).Pluck("name", &existingTemplateNames)

	templateExistMap := make(map[string]bool)
	for _, name := range existingTemplateNames {
		templateExistMap[name] = true
	}

	templates := InitTaskTemplates()
	var newTemplates []model.Task
	for _, tpl := range templates {
		if templateExistMap[tpl.Name] {
			db.Model(&model.Task{}).Where("name = ?", tpl.Name).
				Updates(map[string]any{
					"type":        tpl.Type,
					"description": tpl.Description,
					"config":      tpl.Config,
					"exec_mode":   tpl.ExecMode,
				})
		} else {
			newTemplates = append(newTemplates, tpl)
		}
	}

	if len(newTemplates) > 0 {
		db.Create(&newTemplates)
	}

	var existingScheduleTypes []string

	db.Model(&model.TaskSchedule{}).
		Where("task_type LIKE ?", "system-%").
		Pluck("task_type", &existingScheduleTypes)

	scheduleExistMap := make(map[string]bool)
	for _, tType := range existingScheduleTypes {
		scheduleExistMap[tType] = true
	}

	schedules := InitTaskSchedule()
	var newSchedules []model.TaskSchedule
	for _, schedule := range schedules {
		if !scheduleExistMap[schedule.TaskType] {
			newSchedules = append(newSchedules, schedule)
		}
	}

	if len(newSchedules) > 0 {
		db.Create(&newSchedules)
	}
}

func InitTaskTemplates() []model.Task {
	templates := []model.Task{
		{
			Type:        "system",
			Name:        "local_cpu",
			Description: "展示cpu占用率",
			Config:      model.TempleConfig(`[]`),
			ExecMode:    "manual",
		},
		{
			Type:        "system",
			Name:        "local_memory",
			Description: "展示内存占用率",
			Config:      model.TempleConfig(`[]`),
			ExecMode:    "manual",
		},
		{
			Type:        "system",
			Name:        "active_tasks",
			Description: "展示任务数",
			Config:      model.TempleConfig(`[]`),
			ExecMode:    "manual",
		},
		{
			Type:        "system",
			Name:        "local_network",
			Description: "本地网络状况",
			Config:      model.TempleConfig(`[]`),
			ExecMode:    "manual",
		},
		{
			Type:        "campus_auth",
			Name:        "校园网自动连",
			Description: "检测网络状态并在掉线时自动执行登录认证",
			Config: model.TempleConfig(`[
				{"field": "address", "label": "地点", "input_type": "select", "options": [{"value": "宜宾", "label": "宜宾"}, {"value": "自贡", "label": "自贡"}]},
				{"field": "service", "label": "服务商", "input_type": "select", "options": [{"value": "移动", "label": "移动"}, {"value": "电信", "label": "电信"}, {"value": "联通", "label": "联通"}]},
				{"field": "username", "label": "校园网账号", "input_type": "text", "placeholder": "请输入学号"},
				{"field": "password", "label": "密码", "input_type": "password", "placeholder": "请输入密码"}
			]`),
			ExecMode: "both",
		},
		{
			Type:        "652_signin",
			Name:        "652 自动签到",
			Description: "CAS 统一身份认证自动登录并提交签到定位",
			Config: model.TempleConfig(`[
				{"field": "account", "label": "学工号", "input_type": "text", "placeholder": "请输入学工号"},
				{"field": "password", "label": "密码", "input_type": "password", "placeholder": "请输入密码"},
				{"field": "local", "label": "校区", "input_type": "select", "options": [{"value": "宜宾", "label": "宜宾"}, {"value": "李白河", "label": "李白河"}, {"value": "汇东", "label": "汇东"}]}
			]`),
			ExecMode: "both",
		},
		{
			Type:        "load_test",
			Name:        "性能压测",
			Description: "对指定目标进行高并发HTTP压力测试",
			Config: model.TempleConfig(`[
				{"field": "url", "label": "压测目标 URL", "input_type": "text", "placeholder": "https://api.example.com"},
				{"field": "concurrency", "label": "并发请求数", "input_type": "number", "default_val": "50"},
				{"field": "duration", "label": "持续时间(秒)", "input_type": "number", "default_val": "60"}
			]`),
			ExecMode: "manual",
		},
		{
			Type:        "net_radar",
			Name:        "延迟雷达",
			Description: "实时监控网络延迟和丢包率",
			Config: model.TempleConfig(`[
				{"field": "target", "label": "监控目标 IP/域名", "input_type": "text", "placeholder": "114.114.114.114"},
				{"field": "interval", "label": "探测频率(秒)", "input_type": "number", "default_val": "5"}
			]`),
			ExecMode: "manual",
		},
		{
			Type:        "port_killer",
			Name:        "端口杀手",
			Description: "扫描并一键关闭占用特定端口的系统进程",
			Config: model.TempleConfig(`[
				{"field": "port", "label": "目标端口号", "input_type": "number", "placeholder": "例如: 8080"}
			]`),
			ExecMode: "manual",
		},
	}

	return templates
}

func InitTaskSchedule() []model.TaskSchedule {
	defaultSchedules := []model.TaskSchedule{
		{
			Name:      "本地 CPU 监控",
			TaskType:  "system-local_cpu",
			IsEnabled: true,
			Config:    model.TempleConfig(`{}`),
		},
		{
			Name:      "本地 内存 监控",
			TaskType:  "system-local_memory",
			IsEnabled: true,
			Config:    model.TempleConfig(`{}`),
		},
		{
			Name:      "全局 活跃任务数",
			TaskType:  "system-active_tasks",
			IsEnabled: true,
			Config:    model.TempleConfig(`{}`),
		},
		{
			Name:      "本地 网络 监控",
			TaskType:  "system-local_network",
			IsEnabled: true,
			Config:    model.TempleConfig(`{}`),
		},
	}
	return defaultSchedules
}
