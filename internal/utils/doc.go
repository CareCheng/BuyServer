// Package utils 工具函数包
// 本包提供项目中通用的工具函数。
//
// 主要功能模块：
//
// 1. 加密相关 (crypto.go)
//   - HashPassword: 密码哈希
//   - CheckPassword: 密码验证
//   - GenerateRandomString: 生成随机字符串
//   - AESEncrypt/AESDecrypt: AES 加密解密
//
// 2. 订单相关 (order.go)
//   - GenerateOrderNo: 生成订单号
//   - GenerateLocalOrderNo: 生成本地订单号
//
// 3. 时间相关 (time.go)
//   - ToDays: 转换为天数
//   - ParseDuration: 解析时间跨度
//
// 4. 验证相关 (validator.go)
//   - ValidateEmail: 验证邮箱格式
//   - ValidatePhone: 验证手机号格式
//   - ValidateUsername: 验证用户名格式
//
// 5. 日志相关 (logger.go)
//   - Logger: 统一日志记录器
//   - LogLevel: 日志级别
//
// 6. 环境配置 (environment.go)
//   - GetEnvironment: 获取运行环境
//   - IsDevelopment: 是否开发环境
//   - IsProduction: 是否生产环境
//
// 使用示例：
//
//	// 密码哈希
//	hash, err := utils.HashPassword("password123")
//
//	// 生成订单号
//	orderNo := utils.GenerateOrderNo(false)
//
//	// 验证邮箱
//	if utils.ValidateEmail("user@example.com") {
//	    // 邮箱格式正确
//	}
package utils

// ==================== 工具函数设计原则 ====================
//
// 1. 无副作用
//    - 函数不修改全局状态
//    - 输入相同则输出相同
//
// 2. 错误处理
//    - 可能失败的操作返回 error
//    - 简单验证返回 bool
//
// 3. 参数验证
//    - 在函数开头验证参数
//    - 无效参数返回明确错误
//
// 4. 性能考虑
//    - 避免不必要的内存分配
//    - 复用缓冲区
//    - 使用 sync.Pool 缓存对象
//
// ==================== 加密安全说明 ====================
//
// 密码哈希：
//   - 使用 bcrypt 算法
//   - cost 参数设置为 10
//   - 每次生成不同的盐值
//
// AES 加密：
//   - 使用 AES-256-GCM 模式
//   - 密钥长度 32 字节
//   - 随机生成 IV
//
// 随机数生成：
//   - 使用 crypto/rand 生成安全随机数
//   - 不使用 math/rand 生成敏感数据
//
// ==================== 订单号格式说明 ====================
//
// 订单号格式：
//   [时间戳]-[随机数]
//   例如：20240101120000-A1B2C3
//
// 本地卡密模式订单号格式：
//   LOCAL_[时间戳]_[随机数]
//   例如：LOCAL_20240101120000_X1Y2
//
// 测试订单前缀：TEST_
//
// ==================== 日志级别说明 ====================
//
// Debug: 调试信息，仅开发环境使用
// Info: 一般信息，记录正常操作
// Warn: 警告信息，需要关注但不影响运行
// Error: 错误信息，影响功能但系统可继续运行
// Fatal: 致命错误，系统无法继续运行
