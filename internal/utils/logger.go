package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ==================== 日志级别定义 ====================

// LogLevel 日志级别
type LogLevel int

const (
	// LogLevelDebug 调试级别
	LogLevelDebug LogLevel = iota
	// LogLevelInfo 信息级别
	LogLevelInfo
	// LogLevelWarn 警告级别
	LogLevelWarn
	// LogLevelError 错误级别
	LogLevelError
	// LogLevelFatal 致命错误级别
	LogLevelFatal
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ColorString 返回带颜色的日志级别字符串
func (l LogLevel) ColorString() string {
	switch l {
	case LogLevelDebug:
		return "\033[36mDEBUG\033[0m" // 青色
	case LogLevelInfo:
		return "\033[32mINFO\033[0m" // 绿色
	case LogLevelWarn:
		return "\033[33mWARN\033[0m" // 黄色
	case LogLevelError:
		return "\033[31mERROR\033[0m" // 红色
	case LogLevelFatal:
		return "\033[35mFATAL\033[0m" // 紫色
	default:
		return "UNKNOWN"
	}
}

// ==================== 日志条目结构 ====================

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Caller    string                 `json:"caller,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// ==================== 日志配置 ====================

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level       LogLevel // 最低日志级别
	Format      string   // 输出格式: "text" 或 "json"
	Output      string   // 输出目标: "stdout", "file", "both"
	FilePath    string   // 日志文件路径
	MaxSize     int      // 单个日志文件最大大小（MB）
	MaxBackups  int      // 保留的旧日志文件数量
	MaxAge      int      // 日志文件保留天数
	Compress    bool     // 是否压缩旧日志
	EnableColor bool     // 是否启用颜色（仅对stdout有效）
	EnableCaller bool    // 是否显示调用位置
}

// DefaultLoggerConfig 默认日志配置
var DefaultLoggerConfig = LoggerConfig{
	Level:       LogLevelInfo,
	Format:      "text",
	Output:      "stdout",
	FilePath:    "logs/app.log",
	MaxSize:     100,
	MaxBackups:  3,
	MaxAge:      7,
	Compress:    false,
	EnableColor: true,
	EnableCaller: true,
}

// ==================== 日志记录器 ====================

// Logger 日志记录器
type Logger struct {
	config  LoggerConfig
	mu      sync.Mutex
	writers []io.Writer
	file    *os.File
}

// 全局日志实例
var (
	defaultLogger *Logger
	loggerOnce    sync.Once
)

// GetLogger 获取全局日志实例
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		defaultLogger = NewLogger(DefaultLoggerConfig)
	})
	return defaultLogger
}

// SetLogger 设置全局日志实例
func SetLogger(logger *Logger) {
	defaultLogger = logger
}

// NewLogger 创建新的日志记录器
func NewLogger(config LoggerConfig) *Logger {
	logger := &Logger{
		config:  config,
		writers: make([]io.Writer, 0),
	}

	// 配置输出目标
	switch config.Output {
	case "stdout":
		logger.writers = append(logger.writers, os.Stdout)
	case "file":
		if f := logger.openLogFile(); f != nil {
			logger.writers = append(logger.writers, f)
		}
	case "both":
		logger.writers = append(logger.writers, os.Stdout)
		if f := logger.openLogFile(); f != nil {
			logger.writers = append(logger.writers, f)
		}
	default:
		logger.writers = append(logger.writers, os.Stdout)
	}

	return logger
}

// openLogFile 打开日志文件
func (l *Logger) openLogFile() *os.File {
	// 确保目录存在
	dir := filepath.Dir(l.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
		return nil
	}

	// 打开或创建日志文件
	file, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开日志文件失败: %v\n", err)
		return nil
	}

	l.file = file
	return file
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// log 记录日志的核心方法
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}, err error) {
	// 检查日志级别
	if level < l.config.Level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 构建日志条目
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05.000"),
		Level:     level.String(),
		Message:   msg,
		Fields:    fields,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// 添加调用位置
	if l.config.EnableCaller {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			// 只保留文件名和行号
			entry.Caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		}
	}

	// 格式化输出
	var output string
	if l.config.Format == "json" {
		data, _ := json.Marshal(entry)
		output = string(data) + "\n"
	} else {
		output = l.formatText(entry, level)
	}

	// 写入所有输出目标
	for _, writer := range l.writers {
		writer.Write([]byte(output))
	}

	// 如果是致命错误，退出程序
	if level == LogLevelFatal {
		os.Exit(1)
	}
}

// formatText 格式化文本输出
func (l *Logger) formatText(entry LogEntry, level LogLevel) string {
	var builder strings.Builder

	// 时间戳
	builder.WriteString(entry.Timestamp)
	builder.WriteString(" ")

	// 日志级别（可选颜色）
	if l.config.EnableColor && l.config.Output != "file" {
		builder.WriteString(fmt.Sprintf("[%-5s]", level.ColorString()))
	} else {
		builder.WriteString(fmt.Sprintf("[%-5s]", level.String()))
	}
	builder.WriteString(" ")

	// 调用位置
	if entry.Caller != "" {
		builder.WriteString(fmt.Sprintf("(%s) ", entry.Caller))
	}

	// 消息
	builder.WriteString(entry.Message)

	// 字段
	if len(entry.Fields) > 0 {
		builder.WriteString(" | ")
		first := true
		for k, v := range entry.Fields {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	// 错误
	if entry.Error != "" {
		builder.WriteString(" | error=")
		builder.WriteString(entry.Error)
	}

	builder.WriteString("\n")
	return builder.String()
}

// ==================== 日志方法 ====================

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelDebug, msg, f, nil)
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelInfo, msg, f, nil)
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelWarn, msg, f, nil)
}

// Error 记录错误日志
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelError, msg, f, err)
}

// Fatal 记录致命错误日志并退出程序
func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelFatal, msg, f, err)
}

// WithFields 创建带有预设字段的日志记录器
func (l *Logger) WithFields(fields map[string]interface{}) *LoggerWithFields {
	return &LoggerWithFields{
		logger: l,
		fields: fields,
	}
}

// ==================== 带字段的日志记录器 ====================

// LoggerWithFields 带预设字段的日志记录器
type LoggerWithFields struct {
	logger *Logger
	fields map[string]interface{}
}

// mergeFields 合并字段
func (l *LoggerWithFields) mergeFields(extra map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return merged
}

// Debug 记录调试日志
func (l *LoggerWithFields) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = l.mergeFields(fields[0])
	} else {
		f = l.fields
	}
	l.logger.log(LogLevelDebug, msg, f, nil)
}

// Info 记录信息日志
func (l *LoggerWithFields) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = l.mergeFields(fields[0])
	} else {
		f = l.fields
	}
	l.logger.log(LogLevelInfo, msg, f, nil)
}

// Warn 记录警告日志
func (l *LoggerWithFields) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = l.mergeFields(fields[0])
	} else {
		f = l.fields
	}
	l.logger.log(LogLevelWarn, msg, f, nil)
}

// Error 记录错误日志
func (l *LoggerWithFields) Error(msg string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = l.mergeFields(fields[0])
	} else {
		f = l.fields
	}
	l.logger.log(LogLevelError, msg, f, err)
}

// ==================== 全局便捷方法 ====================

// Debug 全局调试日志
func Debug(msg string, fields ...map[string]interface{}) {
	GetLogger().Debug(msg, fields...)
}

// Info 全局信息日志
func Info(msg string, fields ...map[string]interface{}) {
	GetLogger().Info(msg, fields...)
}

// Warn 全局警告日志
func Warn(msg string, fields ...map[string]interface{}) {
	GetLogger().Warn(msg, fields...)
}

// Error 全局错误日志
func Error(msg string, err error, fields ...map[string]interface{}) {
	GetLogger().Error(msg, err, fields...)
}

// Fatal 全局致命错误日志
func Fatal(msg string, err error, fields ...map[string]interface{}) {
	GetLogger().Fatal(msg, err, fields...)
}

// WithFields 创建带字段的日志记录器
func WithFields(fields map[string]interface{}) *LoggerWithFields {
	return GetLogger().WithFields(fields)
}

// ==================== 请求日志中间件辅助 ====================

// RequestLogFields 请求日志字段
type RequestLogFields struct {
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	IP         string        `json:"ip"`
	UserAgent  string        `json:"user_agent"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency"`
	UserID     uint          `json:"user_id,omitempty"`
}

// ToMap 转换为map
func (r *RequestLogFields) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"method":      r.Method,
		"path":        r.Path,
		"ip":          r.IP,
		"user_agent":  r.UserAgent,
		"status_code": r.StatusCode,
		"latency_ms":  r.Latency.Milliseconds(),
	}
	if r.UserID > 0 {
		m["user_id"] = r.UserID
	}
	return m
}

// LogRequest 记录请求日志
func LogRequest(fields *RequestLogFields) {
	level := LogLevelInfo
	if fields.StatusCode >= 500 {
		level = LogLevelError
	} else if fields.StatusCode >= 400 {
		level = LogLevelWarn
	}

	msg := fmt.Sprintf("%s %s %d %dms", fields.Method, fields.Path, fields.StatusCode, fields.Latency.Milliseconds())
	GetLogger().log(level, msg, fields.ToMap(), nil)
}
