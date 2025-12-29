package model

import (
	"time"
)

// AdminRole 管理员角色
type AdminRole struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex" json:"name"`                // 角色名称
	Description string    `gorm:"size:200" json:"description"`                    // 角色描述
	Permissions string    `gorm:"type:text" json:"permissions"`                   // JSON格式权限列表
	IsSystem    bool      `gorm:"default:false" json:"is_system"`                 // 是否系统内置角色
	Status      int       `gorm:"default:1" json:"status"`                        // 状态：1启用 0禁用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Admin 管理员账户
type Admin struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex" json:"username"`          // 用户名
	PasswordHash string     `gorm:"size:255" json:"-"`                            // 密码哈希
	RoleID       uint       `gorm:"index" json:"role_id"`                         // 角色ID
	Role         *AdminRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`      // 角色关联
	Email        string     `gorm:"size:255" json:"email"`                        // 邮箱
	Nickname     string     `gorm:"size:100" json:"nickname"`                     // 昵称
	Avatar       string     `gorm:"size:500" json:"avatar"`                       // 头像
	Enable2FA    bool       `gorm:"default:false" json:"enable_2fa"`              // 是否启用两步验证
	TOTPSecret   string     `gorm:"size:64" json:"-"`                             // TOTP密钥
	Status       int        `gorm:"default:1" json:"status"`                      // 状态：1启用 0禁用
	LastLoginAt  *time.Time `json:"last_login_at"`                                // 最后登录时间
	LastLoginIP  string     `gorm:"size:50" json:"last_login_ip"`                 // 最后登录IP
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Permission 权限定义
type Permission struct {
	Code        string `json:"code"`        // 权限代码
	Name        string `json:"name"`        // 权限名称
	Description string `json:"description"` // 权限描述
	Group       string `json:"group"`       // 权限分组
}

// 预定义权限列表
var AllPermissions = []Permission{
	// 仪表盘
	{Code: "dashboard:view", Name: "查看仪表盘", Description: "查看系统仪表盘统计数据", Group: "仪表盘"},

	// 商品管理
	{Code: "product:view", Name: "查看商品", Description: "查看商品列表和详情", Group: "商品管理"},
	{Code: "product:create", Name: "创建商品", Description: "创建新商品", Group: "商品管理"},
	{Code: "product:edit", Name: "编辑商品", Description: "编辑商品信息", Group: "商品管理"},
	{Code: "product:delete", Name: "删除商品", Description: "删除商品", Group: "商品管理"},

	// 订单管理
	{Code: "order:view", Name: "查看订单", Description: "查看订单列表和详情", Group: "订单管理"},
	{Code: "order:edit", Name: "编辑订单", Description: "编辑订单状态", Group: "订单管理"},
	{Code: "order:delete", Name: "删除订单", Description: "删除订单", Group: "订单管理"},
	{Code: "order:export", Name: "导出订单", Description: "导出订单数据", Group: "订单管理"},

	// 用户管理
	{Code: "user:view", Name: "查看用户", Description: "查看用户列表和详情", Group: "用户管理"},
	{Code: "user:edit", Name: "编辑用户", Description: "编辑用户信息和状态", Group: "用户管理"},
	{Code: "user:delete", Name: "删除用户", Description: "删除用户", Group: "用户管理"},
	{Code: "user:export", Name: "导出用户", Description: "导出用户数据", Group: "用户管理"},

	// 优惠券管理
	{Code: "coupon:view", Name: "查看优惠券", Description: "查看优惠券列表", Group: "优惠券管理"},
	{Code: "coupon:create", Name: "创建优惠券", Description: "创建新优惠券", Group: "优惠券管理"},
	{Code: "coupon:edit", Name: "编辑优惠券", Description: "编辑优惠券信息", Group: "优惠券管理"},
	{Code: "coupon:delete", Name: "删除优惠券", Description: "删除优惠券", Group: "优惠券管理"},

	// 公告管理
	{Code: "announcement:view", Name: "查看公告", Description: "查看公告列表", Group: "公告管理"},
	{Code: "announcement:create", Name: "创建公告", Description: "创建新公告", Group: "公告管理"},
	{Code: "announcement:edit", Name: "编辑公告", Description: "编辑公告内容", Group: "公告管理"},
	{Code: "announcement:delete", Name: "删除公告", Description: "删除公告", Group: "公告管理"},

	// 分类管理
	{Code: "category:view", Name: "查看分类", Description: "查看分类列表", Group: "分类管理"},
	{Code: "category:create", Name: "创建分类", Description: "创建新分类", Group: "分类管理"},
	{Code: "category:edit", Name: "编辑分类", Description: "编辑分类信息", Group: "分类管理"},
	{Code: "category:delete", Name: "删除分类", Description: "删除分类", Group: "分类管理"},

	// 客服管理
	{Code: "support:view", Name: "查看客服", Description: "查看客服工单和聊天", Group: "客服管理"},
	{Code: "support:manage", Name: "管理客服", Description: "管理客服人员", Group: "客服管理"},
	{Code: "support:config", Name: "客服配置", Description: "配置客服系统", Group: "客服管理"},

	// FAQ管理
	{Code: "faq:view", Name: "查看FAQ", Description: "查看FAQ列表", Group: "FAQ管理"},
	{Code: "faq:create", Name: "创建FAQ", Description: "创建新FAQ", Group: "FAQ管理"},
	{Code: "faq:edit", Name: "编辑FAQ", Description: "编辑FAQ内容", Group: "FAQ管理"},
	{Code: "faq:delete", Name: "删除FAQ", Description: "删除FAQ", Group: "FAQ管理"},

	// 知识库管理
	{Code: "knowledge:view", Name: "查看知识库", Description: "查看知识库文章", Group: "知识库管理"},
	{Code: "knowledge:create", Name: "创建文章", Description: "创建知识库文章", Group: "知识库管理"},
	{Code: "knowledge:edit", Name: "编辑文章", Description: "编辑知识库文章", Group: "知识库管理"},
	{Code: "knowledge:delete", Name: "删除文章", Description: "删除知识库文章", Group: "知识库管理"},

	// 评价管理
	{Code: "review:view", Name: "查看评价", Description: "查看商品评价", Group: "评价管理"},
	{Code: "review:reply", Name: "回复评价", Description: "回复商品评价", Group: "评价管理"},
	{Code: "review:delete", Name: "删除评价", Description: "删除商品评价", Group: "评价管理"},

	// 系统设置
	{Code: "settings:view", Name: "查看设置", Description: "查看系统设置", Group: "系统设置"},
	{Code: "settings:edit", Name: "编辑设置", Description: "编辑系统设置", Group: "系统设置"},
	{Code: "settings:payment", Name: "支付配置", Description: "配置支付方式", Group: "系统设置"},
	{Code: "settings:email", Name: "邮箱配置", Description: "配置邮箱服务", Group: "系统设置"},
	{Code: "settings:database", Name: "数据库配置", Description: "配置数据库", Group: "系统设置"},
	{Code: "settings:security", Name: "安全设置", Description: "配置安全选项", Group: "系统设置"},

	// 日志管理
	{Code: "log:view", Name: "查看日志", Description: "查看操作日志", Group: "日志管理"},
	{Code: "log:export", Name: "导出日志", Description: "导出操作日志", Group: "日志管理"},

	// 备份管理
	{Code: "backup:view", Name: "查看备份", Description: "查看备份列表", Group: "备份管理"},
	{Code: "backup:create", Name: "创建备份", Description: "创建数据库备份", Group: "备份管理"},
	{Code: "backup:download", Name: "下载备份", Description: "下载备份文件", Group: "备份管理"},
	{Code: "backup:delete", Name: "删除备份", Description: "删除备份文件", Group: "备份管理"},

	// 统计报表
	{Code: "stats:view", Name: "查看统计", Description: "查看统计报表", Group: "统计报表"},
	{Code: "stats:export", Name: "导出统计", Description: "导出统计数据", Group: "统计报表"},

	// 系统监控
	{Code: "monitor:view", Name: "查看监控", Description: "查看系统监控", Group: "系统监控"},

	// 管理员管理
	{Code: "admin:view", Name: "查看管理员", Description: "查看管理员列表", Group: "管理员管理"},
	{Code: "admin:create", Name: "创建管理员", Description: "创建新管理员", Group: "管理员管理"},
	{Code: "admin:edit", Name: "编辑管理员", Description: "编辑管理员信息", Group: "管理员管理"},
	{Code: "admin:delete", Name: "删除管理员", Description: "删除管理员", Group: "管理员管理"},

	// 角色管理
	{Code: "role:view", Name: "查看角色", Description: "查看角色列表", Group: "角色管理"},
	{Code: "role:create", Name: "创建角色", Description: "创建新角色", Group: "角色管理"},
	{Code: "role:edit", Name: "编辑角色", Description: "编辑角色权限", Group: "角色管理"},
	{Code: "role:delete", Name: "删除角色", Description: "删除角色", Group: "角色管理"},
}

// GetPermissionGroups 获取按分组整理的权限列表
func GetPermissionGroups() map[string][]Permission {
	groups := make(map[string][]Permission)
	for _, p := range AllPermissions {
		groups[p.Group] = append(groups[p.Group], p)
	}
	return groups
}

// PermissionTemplate 权限模板
type PermissionTemplate struct {
	Name        string   `json:"name"`        // 模板名称
	Description string   `json:"description"` // 模板描述
	Permissions []string `json:"permissions"` // 权限代码列表
}

// 预定义权限模板
var PermissionTemplates = []PermissionTemplate{
	{
		Name:        "viewer",
		Description: "只读用户 - 只能查看数据，无法进行任何修改操作",
		Permissions: []string{
			"dashboard:view",
			"product:view",
			"order:view",
			"user:view",
			"coupon:view",
			"announcement:view",
			"category:view",
			"support:view",
			"faq:view",
			"knowledge:view",
			"review:view",
			"stats:view",
			"log:view",
		},
	},
	{
		Name:        "support",
		Description: "客服人员 - 处理用户咨询、工单和订单查询",
		Permissions: []string{
			"dashboard:view",
			"order:view",
			"user:view",
			"support:view",
			"faq:view",
			"knowledge:view",
			"review:view",
			"review:reply",
		},
	},
	{
		Name:        "operator",
		Description: "运营人员 - 管理商品、内容、优惠券和公告",
		Permissions: []string{
			"dashboard:view",
			"product:view",
			"product:create",
			"product:edit",
			"order:view",
			"coupon:view",
			"coupon:create",
			"coupon:edit",
			"announcement:view",
			"announcement:create",
			"announcement:edit",
			"category:view",
			"category:create",
			"category:edit",
			"faq:view",
			"faq:create",
			"faq:edit",
			"review:view",
			"review:reply",
			"stats:view",
		},
	},
	{
		Name:        "admin",
		Description: "普通管理员 - 拥有大部分管理权限，但无法修改系统配置",
		Permissions: []string{
			"dashboard:view",
			"product:view",
			"product:create",
			"product:edit",
			"product:delete",
			"order:view",
			"order:edit",
			"order:export",
			"user:view",
			"user:edit",
			"coupon:view",
			"coupon:create",
			"coupon:edit",
			"coupon:delete",
			"announcement:view",
			"announcement:create",
			"announcement:edit",
			"announcement:delete",
			"category:view",
			"category:create",
			"category:edit",
			"category:delete",
			"support:view",
			"support:manage",
			"faq:view",
			"faq:create",
			"faq:edit",
			"faq:delete",
			"knowledge:view",
			"knowledge:create",
			"knowledge:edit",
			"knowledge:delete",
			"review:view",
			"review:reply",
			"review:delete",
			"settings:view",
			"log:view",
			"backup:view",
			"backup:create",
			"stats:view",
			"stats:export",
			"monitor:view",
		},
	},
	{
		Name:        "super_admin",
		Description: "超级管理员 - 拥有所有权限，包括系统配置和管理员管理",
		Permissions: func() []string {
			perms := make([]string, 0, len(AllPermissions))
			for _, p := range AllPermissions {
				perms = append(perms, p.Code)
			}
			return perms
		}(),
	},
}

// GetPermissionTemplate 根据名称获取权限模板
func GetPermissionTemplate(name string) *PermissionTemplate {
	for _, t := range PermissionTemplates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// TableName 设置表名
func (AdminRole) TableName() string {
	return "admin_roles"
}

// TableName 设置表名
func (Admin) TableName() string {
	return "admins"
}
