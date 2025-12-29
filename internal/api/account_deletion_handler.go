package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// RequestAccountDeletion 申请账户注销
// POST /api/user/account/delete
func RequestAccountDeletion(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if AccountDeletionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	// 获取用户邮箱
	var email string
	if UserSvc != nil {
		if user, err := UserSvc.GetUserByID(userID); err == nil {
			email = user.Email
		}
	}

	request, err := AccountDeletionSvc.RequestDeletion(userID, username, email, req.Reason)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "request_deletion", "account", strconv.Itoa(int(userID)), "申请账户注销", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "注销申请已提交，我们将在7个工作日内处理",
		"data":    request,
	})
}

// CancelAccountDeletion 取消账户注销申请
// POST /api/user/account/delete/cancel
func CancelAccountDeletion(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	if AccountDeletionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := AccountDeletionSvc.CancelDeletion(userID); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "cancel_deletion", "account", strconv.Itoa(int(userID)), "取消账户注销申请", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "注销申请已取消"})
}

// GetAccountDeletionStatus 获取账户注销状态
// GET /api/user/account/delete/status
func GetAccountDeletionStatus(c *gin.Context) {
	userID := c.GetUint("user_id")

	if AccountDeletionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	request, err := AccountDeletionSvc.GetUserDeletionRequest(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取状态失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"has_request": request != nil,
			"request":     request,
		},
	})
}

// AdminGetDeletionRequests 管理员获取注销申请列表
// GET /api/admin/account/deletions
func AdminGetDeletionRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if AccountDeletionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	requests, total, err := AccountDeletionSvc.GetAllRequests(page, pageSize, status)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取列表失败"})
		return
	}

	// 返回 deletions 字段以匹配前端期望
	c.JSON(200, gin.H{
		"success":   true,
		"deletions": requests,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdminApproveDeletion 管理员批准注销申请
// POST /api/admin/account/deletion/:id/approve
func AdminApproveDeletion(c *gin.Context) {
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "申请ID无效"})
		return
	}

	adminUsername := c.GetString("admin_username")

	if AccountDeletionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := AccountDeletionSvc.ApproveDeletion(uint(requestID), adminUsername); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "approve_deletion", "account_deletion", strconv.Itoa(int(requestID)), "批准账户注销申请", c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "已批准注销申请，账户将在7天后删除"})
}

// AdminRejectDeletion 管理员拒绝注销申请
// POST /api/admin/account/deletion/:id/reject
func AdminRejectDeletion(c *gin.Context) {
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "申请ID无效"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	adminUsername := c.GetString("admin_username")

	if AccountDeletionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	if err := AccountDeletionSvc.RejectDeletion(uint(requestID), adminUsername, req.Reason); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername, "reject_deletion", "account_deletion", strconv.Itoa(int(requestID)), "拒绝账户注销申请: "+req.Reason, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "已拒绝注销申请"})
}
