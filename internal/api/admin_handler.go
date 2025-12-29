package api

import (
	"os"
	"path/filepath"
	"strings"
)

// ==================== 管理后台工具函数 ====================
// 注意：管理后台的 API 处理函数已拆分到以下文件：
// - admin_auth_handler.go      管理员认证（登录、TOTP、登出、2FA）
// - admin_dashboard_handler.go 仪表盘
// - admin_product_handler.go   商品管理
// - admin_order_handler.go     订单管理
// - admin_user_handler.go      用户管理
// - admin_settings_handler.go  系统设置
// - admin_db_handler.go        数据库配置
// - admin_payment_config_handler.go 支付配置
// - admin_email_handler.go     邮箱配置

// getExecDir 获取程序所在目录
func getExecDir() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(execPath)
}

// toRelativePath 将完整路径转换为相对于程序目录的相对路径
func toRelativePath(fullPath string) string {
	if fullPath == "" {
		return fullPath
	}
	execDir := getExecDir()
	if execDir == "" {
		return fullPath
	}
	// 尝试获取相对路径
	relPath, err := filepath.Rel(execDir, fullPath)
	if err != nil {
		return fullPath
	}
	// 将反斜杠转换为正斜杠（跨平台兼容）
	return strings.ReplaceAll(relPath, "\\", "/")
}

// toAbsolutePath 将相对路径转换为完整路径
func toAbsolutePath(relPath string) string {
	if relPath == "" {
		return relPath
	}
	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(relPath) {
		return relPath
	}
	execDir := getExecDir()
	if execDir == "" {
		return relPath
	}
	return filepath.Join(execDir, relPath)
}
