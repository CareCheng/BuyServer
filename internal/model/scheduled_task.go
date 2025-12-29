package model

import "time"

// ScheduledTask 定时任务模型
// 用于管理系统自动化任务，如定时发送报表、清理过期数据等
type ScheduledTask struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:100;not null" json:"name"`        // 任务名称
	Type        string     `gorm:"size:50;not null" json:"type"`         // 任务类型
	CronExpr    string     `gorm:"size:50" json:"cron_expr"`             // Cron表达式
	Config      string     `gorm:"type:text" json:"config"`              // 任务配置（JSON格式）
	Status      int        `gorm:"default:1" json:"status"`              // 状态：1启用 0禁用
	LastRunAt   *time.Time `json:"last_run_at"`                          // 上次执行时间
	NextRunAt   *time.Time `json:"next_run_at"`                          // 下次执行时间
	LastResult  string     `gorm:"size:255" json:"last_result"`          // 上次执行结果
	RunCount    int        `gorm:"default:0" json:"run_count"`           // 执行次数
	FailCount   int        `gorm:"default:0" json:"fail_count"`          // 失败次数
	Description string     `gorm:"size:500" json:"description"`          // 任务描述
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

// TaskLog 任务执行日志
type TaskLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"index;not null" json:"task_id"`        // 任务ID
	TaskName  string    `gorm:"size:100" json:"task_name"`            // 任务名称
	Status    string    `gorm:"size:20" json:"status"`                // 执行状态：success/failed
	Duration  int       `gorm:"default:0" json:"duration"`            // 执行耗时（毫秒）
	Result    string    `gorm:"type:text" json:"result"`              // 执行结果
	Error     string    `gorm:"type:text" json:"error"`               // 错误信息
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (TaskLog) TableName() string {
	return "task_logs"
}

// 任务类型常量
const (
	TaskTypeCleanExpiredOrders  = "clean_expired_orders"  // 清理过期订单
	TaskTypeCleanExpiredSessions = "clean_expired_sessions" // 清理过期会话
	TaskTypeSendDailyReport     = "send_daily_report"     // 发送每日报表
	TaskTypeSendWeeklyReport    = "send_weekly_report"    // 发送每周报表
	TaskTypeBackupDatabase      = "backup_database"       // 数据库备份
	TaskTypeCleanOldLogs        = "clean_old_logs"        // 清理旧日志
	TaskTypeExpirePoints        = "expire_points"         // 积分过期处理
)
