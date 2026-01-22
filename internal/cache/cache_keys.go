// Package cache 提供统一的缓存抽象层
// cache_keys.go - 缓存键定义和生成
package cache

import (
	"fmt"
	"time"
)

// 键前缀（可配置）
var keyPrefix = "user:"

// SetKeyPrefix 设置键前缀
func SetKeyPrefix(prefix string) {
	keyPrefix = prefix
}

// GetKeyPrefix 获取键前缀
func GetKeyPrefix() string {
	return keyPrefix
}

// ==================== 缓存键前缀常量 ====================

const (
	PrefixSession      = "session:"       // 用户会话
	PrefixAdminSession = "admin_session:" // 管理员会话
	PrefixUser         = "user:"          // 用户信息
	PrefixProduct      = "product:"       // 商品
	PrefixCategory     = "category:"      // 分类
	PrefixCart         = "cart:"          // 购物车
	PrefixCoupon       = "coupon:"        // 优惠券
	PrefixPoints       = "points:"        // 积分
	PrefixFavorite     = "favorite:"      // 收藏
	PrefixRate         = "rate:"          // 限流
	PrefixEmailCode    = "email_code:"    // 邮箱验证码
	PrefixConfig       = "config:"        // 系统配置
	PrefixAnnounce     = "announce:"      // 公告
	PrefixFAQ          = "faq:"           // FAQ
	PrefixKnowledge    = "knowledge:"     // 知识库
	PrefixHomepage     = "homepage:"      // 首页配置
	PrefixLogin        = "login:"         // 登录相关
	PrefixReview       = "review:"        // 评价
	PrefixRecharge     = "recharge:"      // 充值优惠
)

// ==================== 缓存 TTL 常量 ====================

const (
	// 会话相关
	SessionTTL      = 2 * time.Hour      // 用户会话
	RememberMeTTL   = 7 * 24 * time.Hour // 记住我
	AdminSessionTTL = 1 * time.Hour      // 管理员会话

	// 用户相关
	UserInfoTTL = 5 * time.Minute    // 用户基本信息
	CartTTL     = 7 * 24 * time.Hour // 购物车（长期缓存）
	PointsTTL   = 1 * time.Minute    // 积分余额
	FavoriteTTL = 5 * time.Minute    // 收藏列表

	// 商品相关
	ProductTTL     = 5 * time.Minute  // 商品缓存
	CategoryTTL    = 10 * time.Minute // 分类缓存
	ReviewStatsTTL = 5 * time.Minute  // 评价统计

	// 优惠相关
	CouponTTL   = 5 * time.Minute  // 优惠券列表
	RechargeTTL = 10 * time.Minute // 充值优惠规则

	// 内容相关
	ConfigTTL    = 5 * time.Minute  // 配置缓存
	AnnounceTTL  = 5 * time.Minute  // 公告缓存
	FAQTTL       = 10 * time.Minute // FAQ 缓存
	KnowledgeTTL = 10 * time.Minute // 知识库缓存
	HomepageTTL  = 5 * time.Minute  // 首页配置

	// 安全相关
	EmailCodeTTL    = 5 * time.Minute  // 验证码有效期
	RateLimitTTL    = time.Minute      // 限流窗口
	LoginFailureTTL = 15 * time.Minute // 登录失败锁定
	LoginLockTTL    = 30 * time.Minute // 账号锁定时长
)

// ==================== 会话相关 ====================

// UserSessionKey 生成用户会话缓存键
// 格式：{prefix}session:{session_id}
func UserSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s%s", keyPrefix, PrefixSession, sessionID)
}

// AdminSessionKey 生成管理员会话缓存键
// 格式：{prefix}admin_session:{session_id}
func AdminSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s%s", keyPrefix, PrefixAdminSession, sessionID)
}

// ==================== 用户信息相关 ====================

// UserInfoKey 生成用户基本信息缓存键
// 格式：{prefix}user:info:{user_id}
func UserInfoKey(userID uint) string {
	return fmt.Sprintf("%s%sinfo:%d", keyPrefix, PrefixUser, userID)
}

// UserPointsKey 生成用户积分缓存键
// 格式：{prefix}points:{user_id}
func UserPointsKey(userID uint) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, PrefixPoints, userID)
}

// ==================== 商品相关 ====================

// ProductKey 生成商品缓存键
// 格式：{prefix}product:{product_id}
func ProductKey(productID uint) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, PrefixProduct, productID)
}

// ProductListKey 生成商品列表缓存键
// 格式：{prefix}product:list:{page}:{size}:{status}:{category_id}
func ProductListKey(page, size int, onlyActive bool, categoryID uint) string {
	status := "all"
	if onlyActive {
		status = "active"
	}
	return fmt.Sprintf("%s%slist:%d:%d:%s:%d", keyPrefix, PrefixProduct, page, size, status, categoryID)
}

// ProductStockKey 生成商品库存缓存键（短 TTL）
// 格式：{prefix}product:stock:{product_id}
func ProductStockKey(productID uint) string {
	return fmt.Sprintf("%s%sstock:%d", keyPrefix, PrefixProduct, productID)
}

// ==================== 分类相关 ====================

// CategoryKey 生成分类缓存键
// 格式：{prefix}category:{category_id}
func CategoryKey(categoryID uint) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, PrefixCategory, categoryID)
}

// CategoryListKey 生成分类列表缓存键
// 格式：{prefix}category:list
func CategoryListKey() string {
	return fmt.Sprintf("%s%slist", keyPrefix, PrefixCategory)
}

// CategoryTreeKey 生成分类树缓存键
// 格式：{prefix}category:tree
func CategoryTreeKey() string {
	return fmt.Sprintf("%s%stree", keyPrefix, PrefixCategory)
}

// ==================== 购物车相关 ====================

// CartKey 生成购物车缓存键
// 格式：{prefix}cart:{user_id}
func CartKey(userID uint) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, PrefixCart, userID)
}

// ==================== 优惠券相关 ====================

// UserCouponsKey 生成用户优惠券列表缓存键
// 格式：{prefix}coupon:user:{user_id}
func UserCouponsKey(userID uint) string {
	return fmt.Sprintf("%s%suser:%d", keyPrefix, PrefixCoupon, userID)
}

// AvailableCouponsKey 生成可用优惠券列表缓存键
// 格式：{prefix}coupon:available
func AvailableCouponsKey() string {
	return fmt.Sprintf("%s%savailable", keyPrefix, PrefixCoupon)
}

// ==================== 收藏相关 ====================

// UserFavoritesKey 生成用户收藏列表缓存键
// 格式：{prefix}favorite:{user_id}
func UserFavoritesKey(userID uint) string {
	return fmt.Sprintf("%s%s%d", keyPrefix, PrefixFavorite, userID)
}

// ==================== 限流相关 ====================

// RateLimitKey 生成限流计数器缓存键
// 格式：{prefix}rate:{type}:{identifier}:{window}
func RateLimitKey(limitType, identifier string, windowID int64) string {
	return fmt.Sprintf("%s%s%s:%s:%d", keyPrefix, PrefixRate, limitType, identifier, windowID)
}

// LoginFailureKey 生成登录失败计数缓存键
// 格式：{prefix}login:failure:{identifier}
func LoginFailureKey(identifier string) string {
	return fmt.Sprintf("%s%sfailure:%s", keyPrefix, PrefixLogin, identifier)
}

// LoginLockKey 生成登录锁定缓存键
// 格式：{prefix}login:lock:{identifier}
func LoginLockKey(identifier string) string {
	return fmt.Sprintf("%s%slock:%s", keyPrefix, PrefixLogin, identifier)
}

// ==================== 验证码相关 ====================

// EmailCodeKey 生成邮箱验证码缓存键
// 格式：{prefix}email_code:{email}:{purpose}
func EmailCodeKey(email, purpose string) string {
	return fmt.Sprintf("%s%s%s:%s", keyPrefix, PrefixEmailCode, email, purpose)
}

// ==================== 配置相关 ====================

// SystemConfigKey 生成系统配置缓存键
// 格式：{prefix}config:{config_type}
func SystemConfigKey(configType string) string {
	return fmt.Sprintf("%s%s%s", keyPrefix, PrefixConfig, configType)
}

// ==================== 公告相关 ====================

// AnnouncementListKey 生成公告列表缓存键
// 格式：{prefix}announce:list
func AnnouncementListKey() string {
	return fmt.Sprintf("%s%slist", keyPrefix, PrefixAnnounce)
}

// ==================== FAQ 相关 ====================

// FAQListKey 生成 FAQ 列表缓存键
// 格式：{prefix}faq:list
func FAQListKey() string {
	return fmt.Sprintf("%s%slist", keyPrefix, PrefixFAQ)
}

// ==================== 知识库相关 ====================

// KnowledgeCategoryKey 生成知识库分类缓存键
// 格式：{prefix}knowledge:category:{category_id}
func KnowledgeCategoryKey(categoryID uint) string {
	return fmt.Sprintf("%s%scategory:%d", keyPrefix, PrefixKnowledge, categoryID)
}

// KnowledgeArticleKey 生成知识库文章缓存键
// 格式：{prefix}knowledge:article:{article_id}
func KnowledgeArticleKey(articleID uint) string {
	return fmt.Sprintf("%s%sarticle:%d", keyPrefix, PrefixKnowledge, articleID)
}

// KnowledgeHotKey 生成热门文章缓存键
// 格式：{prefix}knowledge:hot
func KnowledgeHotKey() string {
	return fmt.Sprintf("%s%shot", keyPrefix, PrefixKnowledge)
}

// ==================== 首页相关 ====================

// HomepageConfigKey 生成首页配置缓存键
// 格式：{prefix}homepage:config
func HomepageConfigKey() string {
	return fmt.Sprintf("%s%sconfig", keyPrefix, PrefixHomepage)
}

// ==================== 评价相关 ====================

// ProductReviewStatsKey 生成商品评价统计缓存键
// 格式：{prefix}review:stats:{product_id}
func ProductReviewStatsKey(productID uint) string {
	return fmt.Sprintf("%s%sstats:%d", keyPrefix, PrefixReview, productID)
}

// ==================== 充值优惠相关 ====================

// RechargeRulesKey 生成充值优惠规则缓存键
// 格式：{prefix}recharge:rules
func RechargeRulesKey() string {
	return fmt.Sprintf("%s%srules", keyPrefix, PrefixRecharge)
}
