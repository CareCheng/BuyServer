package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 角色管理 API ====================

// AdminGetRoles 获取所有角色
func AdminGetRoles(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	roles, err := RoleSvc.GetAllRoles()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取角色列表失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": roles})
}

// AdminGetRole 获取角色详情
func AdminGetRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的角色ID"})
		return
	}

	role, err := RoleSvc.GetRoleByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "角色不存在"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": role})
}

// AdminCreateRole 创建角色
func AdminCreateRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	role, err := RoleSvc.CreateRole(req.Name, req.Description, req.Permissions)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "创建角色", "role", "", req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": role})
}

// AdminUpdateRole 更新角色
func AdminUpdateRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的角色ID"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := RoleSvc.UpdateRole(uint(id), req.Name, req.Description, req.Permissions); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "更新角色", "role", c.Param("id"), req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteRole 删除角色
func AdminDeleteRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的角色ID"})
		return
	}

	// 获取角色信息用于日志
	role, _ := RoleSvc.GetRoleByID(uint(id))
	roleName := ""
	if role != nil {
		roleName = role.Name
	}

	if err := RoleSvc.DeleteRole(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "删除角色", "role", c.Param("id"), gin.H{"name": roleName}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminGetPermissions 获取所有权限定义
func AdminGetPermissions(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	permissions := RoleSvc.GetAllPermissions()
	groups := RoleSvc.GetPermissionGroups()

	c.JSON(200, gin.H{
		"success":     true,
		"permissions": permissions,
		"groups":      groups,
		"templates":   model.PermissionTemplates,
	})
}

// ==================== 管理员管理 API ====================

// AdminGetAdmins 获取管理员列表
func AdminGetAdmins(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	admins, total, err := RoleSvc.GetAllAdmins(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取管理员列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    admins,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminGetAdmin 获取管理员详情
func AdminGetAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	admin, err := RoleSvc.GetAdminByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "管理员不存在"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": admin})
}

// AdminCreateAdmin 创建管理员
func AdminCreateAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		RoleID   uint   `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，密码至少6位"})
		return
	}

	admin, err := RoleSvc.CreateAdmin(req.Username, req.Password, req.Email, req.Nickname, req.RoleID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "创建管理员", "admin", "", gin.H{"username": req.Username, "role_id": req.RoleID}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": admin})
}

// AdminUpdateAdmin 更新管理员
func AdminUpdateAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		RoleID   uint   `json:"role_id" binding:"required"`
		Status   int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := RoleSvc.UpdateAdmin(uint(id), req.Email, req.Nickname, req.RoleID, req.Status); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "更新管理员", "admin", c.Param("id"), req, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminUpdateAdminPassword 更新管理员密码
func AdminUpdateAdminPassword(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "密码至少6位"})
		return
	}

	if err := RoleSvc.UpdateAdminPassword(uint(id), req.Password); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "重置密码", "admin", c.Param("id"), nil, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "密码更新成功"})
}

// AdminDeleteAdmin 删除管理员
func AdminDeleteAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	// 获取管理员信息用于日志
	admin, _ := RoleSvc.GetAdminByID(uint(id))
	adminName := ""
	if admin != nil {
		adminName = admin.Username
	}

	if err := RoleSvc.DeleteAdmin(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "删除管理员", "admin", c.Param("id"), gin.H{"username": adminName}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminGetMyPermissions 获取当前管理员的权限
func AdminGetMyPermissions(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	adminUsername, exists := c.Get("admin_username")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "未登录"})
		return
	}

	admin, err := RoleSvc.GetAdminByUsername(adminUsername.(string))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "管理员不存在"})
		return
	}

	permissions, err := RoleSvc.GetAdminPermissions(admin.ID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取权限失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":     true,
		"permissions": permissions,
		"role":        admin.Role,
	})
}

// PermissionRequired 权限检查中间件
func PermissionRequired(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if RoleSvc == nil {
			c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
			c.Abort()
			return
		}

		adminUsername, exists := c.Get("admin_username")
		if !exists {
			c.JSON(401, gin.H{"success": false, "error": "未登录"})
			c.Abort()
			return
		}

		// 获取管理员角色
		adminRole, exists := c.Get("admin_role")
		if exists && adminRole.(string) == "super_admin" {
			// 超级管理员拥有所有权限
			c.Next()
			return
		}

		admin, err := RoleSvc.GetAdminByUsername(adminUsername.(string))
		if err != nil {
			c.JSON(403, gin.H{"success": false, "error": "无权限访问"})
			c.Abort()
			return
		}

		if !RoleSvc.AdminHasPermission(admin.ID, permission) {
			c.JSON(403, gin.H{"success": false, "error": "无权限执行此操作"})
			c.Abort()
			return
		}

		c.Next()
	}
}
