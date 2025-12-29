package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
)

// TaskService 定时任务服务
type TaskService struct {
	repo      *repository.Repository
	running   bool
	stopChan  chan struct{}
	mutex     sync.Mutex
	taskFuncs map[string]TaskFunc
}

// TaskFunc 任务执行函数类型
type TaskFunc func(config string) error

// NewTaskService 创建任务服务实例
func NewTaskService(repo *repository.Repository) *TaskService {
	s := &TaskService{
		repo:      repo,
		stopChan:  make(chan struct{}),
		taskFuncs: make(map[string]TaskFunc),
	}
	// 注册内置任务
	s.registerBuiltinTasks()
	return s
}

// registerBuiltinTasks 注册内置任务
func (s *TaskService) registerBuiltinTasks() {
	s.taskFuncs[model.TaskTypeCleanExpiredOrders] = s.cleanExpiredOrders
	s.taskFuncs[model.TaskTypeCleanExpiredSessions] = s.cleanExpiredSessions
	s.taskFuncs[model.TaskTypeCleanOldLogs] = s.cleanOldLogs
}

// RegisterTask 注册自定义任务
// 参数：
//   - taskType: 任务类型
//   - fn: 任务执行函数
func (s *TaskService) RegisterTask(taskType string, fn TaskFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.taskFuncs[taskType] = fn
}

// Start 启动任务调度器
func (s *TaskService) Start() {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mutex.Unlock()

	go s.runScheduler()
}

// Stop 停止任务调度器
func (s *TaskService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopChan)
}

// runScheduler 运行调度器
func (s *TaskService) runScheduler() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

// checkAndRunTasks 检查并执行到期任务
func (s *TaskService) checkAndRunTasks() {
	var tasks []model.ScheduledTask
	now := time.Now()
	s.repo.GetDB().Where("status = 1 AND (next_run_at IS NULL OR next_run_at <= ?)", now).Find(&tasks)

	for _, task := range tasks {
		go s.executeTask(&task)
	}
}

// executeTask 执行单个任务
func (s *TaskService) executeTask(task *model.ScheduledTask) {
	startTime := time.Now()
	var result string
	var errMsg string
	status := "success"

	// 获取任务执行函数
	fn, exists := s.taskFuncs[task.Type]
	if !exists {
		errMsg = "未知的任务类型"
		status = "failed"
	} else {
		// 执行任务
		if err := fn(task.Config); err != nil {
			errMsg = err.Error()
			status = "failed"
		} else {
			result = "执行成功"
		}
	}

	duration := int(time.Since(startTime).Milliseconds())

	// 记录执行日志
	log := model.TaskLog{
		TaskID:   task.ID,
		TaskName: task.Name,
		Status:   status,
		Duration: duration,
		Result:   result,
		Error:    errMsg,
	}
	s.repo.GetDB().Create(&log)

	// 更新任务状态
	task.LastRunAt = &startTime
	task.RunCount++
	if status == "failed" {
		task.FailCount++
		task.LastResult = "失败: " + errMsg
	} else {
		task.LastResult = "成功"
	}

	// 计算下次执行时间（简单实现，每天执行一次）
	nextRun := startTime.Add(24 * time.Hour)
	task.NextRunAt = &nextRun

	s.repo.GetDB().Save(task)
}

// cleanExpiredOrders 清理过期订单
func (s *TaskService) cleanExpiredOrders(config string) error {
	// 解析配置
	var cfg struct {
		ExpireMinutes int `json:"expire_minutes"`
	}
	cfg.ExpireMinutes = 30 // 默认30分钟
	if config != "" {
		json.Unmarshal([]byte(config), &cfg)
	}

	// 取消超时未支付的订单
	expireTime := time.Now().Add(-time.Duration(cfg.ExpireMinutes) * time.Minute)
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("status = 0 AND created_at < ?", expireTime).
		Update("status", 3) // 3=已取消

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// cleanExpiredSessions 清理过期会话
func (s *TaskService) cleanExpiredSessions(config string) error {
	// 解析配置
	var cfg struct {
		ExpireDays int `json:"expire_days"`
	}
	cfg.ExpireDays = 7 // 默认7天
	if config != "" {
		json.Unmarshal([]byte(config), &cfg)
	}

	// 删除过期的登录设备记录
	expireTime := time.Now().AddDate(0, 0, -cfg.ExpireDays)
	s.repo.GetDB().Where("last_active < ?", expireTime).Delete(&model.LoginDevice{})

	return nil
}

// cleanOldLogs 清理旧日志
func (s *TaskService) cleanOldLogs(config string) error {
	// 解析配置
	var cfg struct {
		RetainDays int `json:"retain_days"`
	}
	cfg.RetainDays = 30 // 默认保留30天
	if config != "" {
		json.Unmarshal([]byte(config), &cfg)
	}

	expireTime := time.Now().AddDate(0, 0, -cfg.RetainDays)

	// 注意：操作日志已改为文件存储，不再从数据库清理
	// 文件日志的清理由日志服务自行管理

	// 清理任务日志
	s.repo.GetDB().Where("created_at < ?", expireTime).Delete(&model.TaskLog{})

	return nil
}

// GetTasks 获取任务列表
// 返回：
//   - 任务列表
//   - 错误信息（如有）
func (s *TaskService) GetTasks() ([]model.ScheduledTask, error) {
	var tasks []model.ScheduledTask
	err := s.repo.GetDB().Order("id ASC").Find(&tasks).Error
	return tasks, err
}

// GetTask 获取单个任务
// 参数：
//   - taskID: 任务ID
// 返回：
//   - 任务信息
//   - 错误信息（如有）
func (s *TaskService) GetTask(taskID uint) (*model.ScheduledTask, error) {
	var task model.ScheduledTask
	err := s.repo.GetDB().First(&task, taskID).Error
	return &task, err
}

// CreateTask 创建任务
// 参数：
//   - task: 任务信息
// 返回：
//   - 错误信息（如有）
func (s *TaskService) CreateTask(task *model.ScheduledTask) error {
	// 验证任务类型
	if _, exists := s.taskFuncs[task.Type]; !exists {
		return errors.New("不支持的任务类型")
	}
	return s.repo.GetDB().Create(task).Error
}

// UpdateTask 更新任务
// 参数：
//   - task: 任务信息
// 返回：
//   - 错误信息（如有）
func (s *TaskService) UpdateTask(task *model.ScheduledTask) error {
	return s.repo.GetDB().Save(task).Error
}

// DeleteTask 删除任务
// 参数：
//   - taskID: 任务ID
// 返回：
//   - 错误信息（如有）
func (s *TaskService) DeleteTask(taskID uint) error {
	return s.repo.GetDB().Delete(&model.ScheduledTask{}, taskID).Error
}

// RunTaskNow 立即执行任务
// 参数：
//   - taskID: 任务ID
// 返回：
//   - 错误信息（如有）
func (s *TaskService) RunTaskNow(taskID uint) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return errors.New("任务不存在")
	}

	go s.executeTask(task)
	return nil
}

// GetTaskLogs 获取任务执行日志
// 参数：
//   - taskID: 任务ID（0表示所有任务）
//   - page: 页码
//   - pageSize: 每页数量
// 返回：
//   - 日志列表
//   - 总数
//   - 错误信息（如有）
func (s *TaskService) GetTaskLogs(taskID uint, page, pageSize int) ([]model.TaskLog, int64, error) {
	var total int64
	query := s.repo.GetDB().Model(&model.TaskLog{})
	if taskID > 0 {
		query = query.Where("task_id = ?", taskID)
	}
	query.Count(&total)

	var logs []model.TaskLog
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

// GetAvailableTaskTypes 获取可用的任务类型
// 返回：
//   - 任务类型列表
func (s *TaskService) GetAvailableTaskTypes() []map[string]string {
	types := []map[string]string{
		{"type": model.TaskTypeCleanExpiredOrders, "name": "清理过期订单", "description": "自动取消超时未支付的订单"},
		{"type": model.TaskTypeCleanExpiredSessions, "name": "清理过期会话", "description": "清理长时间未活动的登录设备"},
		{"type": model.TaskTypeCleanOldLogs, "name": "清理旧日志", "description": "清理超过保留期限的日志记录"},
		{"type": model.TaskTypeSendDailyReport, "name": "发送每日报表", "description": "每日发送销售统计报表"},
		{"type": model.TaskTypeSendWeeklyReport, "name": "发送每周报表", "description": "每周发送销售统计报表"},
		{"type": model.TaskTypeBackupDatabase, "name": "数据库备份", "description": "定时备份数据库"},
		{"type": model.TaskTypeExpirePoints, "name": "积分过期处理", "description": "处理过期的用户积分"},
	}
	return types
}

// TaskStats 任务统计信息
type TaskStats struct {
	TotalTasks    int64 `json:"total_tasks"`
	ActiveTasks   int64 `json:"active_tasks"`
	TotalRuns     int64 `json:"total_runs"`
	SuccessRuns   int64 `json:"success_runs"`
	FailedRuns    int64 `json:"failed_runs"`
	TodayRuns     int64 `json:"today_runs"`
	TodaySuccess  int64 `json:"today_success"`
	TodayFailed   int64 `json:"today_failed"`
}

// GetTaskStats 获取任务统计
// 返回：
//   - 统计信息
func (s *TaskService) GetTaskStats() TaskStats {
	var stats TaskStats

	s.repo.GetDB().Model(&model.ScheduledTask{}).Count(&stats.TotalTasks)
	s.repo.GetDB().Model(&model.ScheduledTask{}).Where("status = 1").Count(&stats.ActiveTasks)
	s.repo.GetDB().Model(&model.TaskLog{}).Count(&stats.TotalRuns)
	s.repo.GetDB().Model(&model.TaskLog{}).Where("status = 'success'").Count(&stats.SuccessRuns)
	s.repo.GetDB().Model(&model.TaskLog{}).Where("status = 'failed'").Count(&stats.FailedRuns)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	s.repo.GetDB().Model(&model.TaskLog{}).Where("created_at >= ?", todayStart).Count(&stats.TodayRuns)
	s.repo.GetDB().Model(&model.TaskLog{}).Where("created_at >= ? AND status = 'success'", todayStart).Count(&stats.TodaySuccess)
	s.repo.GetDB().Model(&model.TaskLog{}).Where("created_at >= ? AND status = 'failed'", todayStart).Count(&stats.TodayFailed)

	return stats
}

// InitDefaultTasks 初始化默认任务
// 在系统首次启动时创建默认的定时任务
func (s *TaskService) InitDefaultTasks() error {
	// 检查是否已有任务
	var count int64
	s.repo.GetDB().Model(&model.ScheduledTask{}).Count(&count)
	if count > 0 {
		return nil // 已有任务，不需要初始化
	}

	// 创建默认任务
	defaultTasks := []model.ScheduledTask{
		{
			Name:        "清理过期订单",
			Type:        model.TaskTypeCleanExpiredOrders,
			CronExpr:    "0 */5 * * *", // 每5分钟
			Config:      `{"expire_minutes": 30}`,
			Status:      1,
			Description: "自动取消超过30分钟未支付的订单",
		},
		{
			Name:        "清理过期会话",
			Type:        model.TaskTypeCleanExpiredSessions,
			CronExpr:    "0 3 * * *", // 每天凌晨3点
			Config:      `{"expire_days": 7}`,
			Status:      1,
			Description: "清理7天未活动的登录设备记录",
		},
		{
			Name:        "清理旧日志",
			Type:        model.TaskTypeCleanOldLogs,
			CronExpr:    "0 4 * * *", // 每天凌晨4点
			Config:      `{"retain_days": 30}`,
			Status:      0, // 默认禁用
			Description: "清理30天前的操作日志",
		},
	}

	for _, task := range defaultTasks {
		if err := s.repo.GetDB().Create(&task).Error; err != nil {
			return fmt.Errorf("创建默认任务失败: %v", err)
		}
	}

	return nil
}
