package model

import "time"

// UserPoints 用户积分模型
// 用户消费可获得积分，积分可兑换优惠券或商品
type UserPoints struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"` // 用户ID
	Points    int       `gorm:"default:0" json:"points"`             // 当前积分
	TotalEarn int       `gorm:"default:0" json:"total_earn"`         // 累计获得积分
	TotalUsed int       `gorm:"default:0" json:"total_used"`         // 累计使用积分
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`    // 更新时间
}

// TableName 指定表名
func (UserPoints) TableName() string {
	return "user_points"
}

// PointsLog 积分变动记录
type PointsLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`       // 用户ID
	Type      string    `gorm:"size:50;not null" json:"type"`        // 类型：earn/use/expire/admin
	Points    int       `gorm:"not null" json:"points"`              // 变动积分（正数增加，负数减少）
	Balance   int       `gorm:"not null" json:"balance"`             // 变动后余额
	OrderNo   string    `gorm:"size:64" json:"order_no"`             // 关联订单号
	Remark    string    `gorm:"size:255" json:"remark"`              // 备注
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`    // 创建时间
}

// TableName 指定表名
func (PointsLog) TableName() string {
	return "points_logs"
}

// PointsRule 积分规则
type PointsRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`       // 规则名称
	Type        string    `gorm:"size:50;not null" json:"type"`        // 规则类型：order/register/daily
	Points      int       `gorm:"default:0" json:"points"`             // 固定积分值
	Ratio       float64   `gorm:"default:0" json:"ratio"`              // 积分比例（如消费1元=10积分，ratio=10）
	MinAmount   float64   `gorm:"default:0" json:"min_amount"`         // 最低消费金额
	MaxPoints   int       `gorm:"default:0" json:"max_points"`         // 单次最高积分（0表示不限）
	Status      int       `gorm:"default:1" json:"status"`             // 状态：1启用 0禁用
	Description string    `gorm:"size:500" json:"description"`         // 规则描述
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (PointsRule) TableName() string {
	return "points_rules"
}

// PointsExchange 积分兑换记录
type PointsExchange struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`       // 用户ID
	Points     int       `gorm:"not null" json:"points"`              // 消耗积分
	Type       string    `gorm:"size:50;not null" json:"type"`        // 兑换类型：coupon/product
	TargetID   uint      `gorm:"not null" json:"target_id"`           // 兑换目标ID（优惠券ID或商品ID）
	TargetName string    `gorm:"size:200" json:"target_name"`         // 兑换目标名称
	Status     int       `gorm:"default:1" json:"status"`             // 状态：1成功 0失败 2已使用
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (PointsExchange) TableName() string {
	return "points_exchanges"
}

// 积分变动类型常量
const (
	PointsTypeEarn   = "earn"   // 获得积分（消费）
	PointsTypeUse    = "use"    // 使用积分（兑换）
	PointsTypeExpire = "expire" // 积分过期
	PointsTypeAdmin  = "admin"  // 管理员调整
)

// 积分规则类型常量
const (
	PointsRuleOrder    = "order"    // 订单消费
	PointsRuleRegister = "register" // 注册奖励
	PointsRuleDaily    = "daily"    // 每日签到
)
