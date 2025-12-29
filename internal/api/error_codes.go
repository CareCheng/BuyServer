package api

import (
	"github.com/gin-gonic/gin"
)

// ==================== 错误码定义 ====================
// 本文件定义了系统中所有的业务错误码
// 错误码规则：
//   - 0: 成功
//   - 1000-1999: 通用错误
//   - 2000-2999: 用户相关错误
//   - 3000-3999: 订单相关错误
//   - 4000-4999: 商品相关错误
//   - 5000-5999: 支付相关错误
//   - 6000-6999: 客服系统错误
//   - 7000-7999: 管理后台错误
//   - 8000-8999: 系统配置错误

// ErrorCode 错误码类型
type ErrorCode int

// 错误码常量定义
const (
	// ==================== 成功 ====================
	CodeSuccess ErrorCode = 0

	// ==================== 通用错误 (1000-1999) ====================
	CodeInternalError        ErrorCode = 1000 // 内部错误
	CodeParamError           ErrorCode = 1001 // 参数错误
	CodeUnauthorized         ErrorCode = 1002 // 未授权
	CodeForbidden            ErrorCode = 1003 // 禁止访问
	CodeNotFound             ErrorCode = 1004 // 资源不存在
	CodeTooManyRequests      ErrorCode = 1005 // 请求过于频繁
	CodeServiceUnavailable   ErrorCode = 1006 // 服务不可用
	CodeDatabaseError        ErrorCode = 1007 // 数据库错误
	CodeValidationError      ErrorCode = 1008 // 验证错误
	CodeConflict             ErrorCode = 1009 // 资源冲突
	CodeCSRFError            ErrorCode = 1010 // CSRF验证失败
	CodeCaptchaError         ErrorCode = 1011 // 验证码错误
	CodeServiceNotInit       ErrorCode = 1012 // 服务未初始化
	CodeDBNotConnected       ErrorCode = 1013 // 数据库未连接

	// ==================== 用户相关错误 (2000-2999) ====================
	CodeUserNotFound           ErrorCode = 2000 // 用户不存在
	CodeUserExists             ErrorCode = 2001 // 用户已存在
	CodePasswordWrong          ErrorCode = 2002 // 密码错误
	CodePasswordTooShort       ErrorCode = 2003 // 密码太短
	CodePasswordNotMatch       ErrorCode = 2004 // 两次密码不一致
	CodeUserDisabled           ErrorCode = 2005 // 用户已禁用
	CodeLoginFailed            ErrorCode = 2006 // 登录失败
	CodeLoginLocked            ErrorCode = 2007 // 登录已锁定
	CodeSessionExpired         ErrorCode = 2008 // 会话已过期
	Code2FARequired            ErrorCode = 2009 // 需要两步验证
	Code2FACodeWrong           ErrorCode = 2010 // 两步验证码错误
	CodeEmailNotVerified       ErrorCode = 2011 // 邮箱未验证
	CodeEmailVerifyFailed      ErrorCode = 2012 // 邮箱验证失败
	CodeEmailCodeWrong         ErrorCode = 2013 // 邮箱验证码错误
	CodeEmailCodeExpired       ErrorCode = 2014 // 邮箱验证码已过期
	CodeEmailExists            ErrorCode = 2015 // 邮箱已被绑定
	CodeEmailSendFailed        ErrorCode = 2016 // 邮件发送失败
	CodeResetTokenInvalid      ErrorCode = 2017 // 重置令牌无效
	CodeResetTokenExpired      ErrorCode = 2018 // 重置令牌已过期
	CodeUsernameInvalid        ErrorCode = 2019 // 用户名格式无效
	CodePhoneInvalid           ErrorCode = 2020 // 手机号格式无效
	CodeDeviceNotFound         ErrorCode = 2021 // 设备不存在
	CodeAccountDeletionPending ErrorCode = 2022 // 账户注销申请处理中
	CodeTOTPNotSet             ErrorCode = 2023 // 未设置动态口令
	CodeTOTPCodeWrong          ErrorCode = 2024 // 动态口令错误

	// ==================== 订单相关错误 (3000-3999) ====================
	CodeOrderNotFound      ErrorCode = 3000 // 订单不存在
	CodeOrderStatusError   ErrorCode = 3001 // 订单状态异常
	CodeOrderExpired       ErrorCode = 3002 // 订单已过期
	CodeOrderCanceled      ErrorCode = 3003 // 订单已取消
	CodeOrderPaid          ErrorCode = 3004 // 订单已支付
	CodeOrderCreateFailed  ErrorCode = 3005 // 订单创建失败
	CodeOrderNotOwner      ErrorCode = 3006 // 不是订单所有者
	CodeOrderCannotCancel  ErrorCode = 3007 // 订单无法取消
	CodeKamiNotAvailable   ErrorCode = 3008 // 卡密不可用
	CodeKamiGetFailed      ErrorCode = 3009 // 获取卡密失败
	CodeInvoiceNotFound    ErrorCode = 3010 // 发票不存在
	CodeInvoiceApplyFailed ErrorCode = 3011 // 发票申请失败

	// ==================== 商品相关错误 (4000-4999) ====================
	CodeProductNotFound   ErrorCode = 4000 // 商品不存在
	CodeProductOffline    ErrorCode = 4001 // 商品已下架
	CodeProductSoldOut    ErrorCode = 4002 // 商品已售罄
	CodeCategoryNotFound  ErrorCode = 4003 // 分类不存在
	CodeProductTestDeny   ErrorCode = 4004 // 商品不支持测试购买
	CodeImageUploadFailed ErrorCode = 4005 // 图片上传失败
	CodeImageNotFound     ErrorCode = 4006 // 图片不存在
	CodeReviewNotFound    ErrorCode = 4007 // 评价不存在
	CodeReviewExists      ErrorCode = 4008 // 已经评价过
	CodeReviewNotAllowed  ErrorCode = 4009 // 不允许评价

	// ==================== 支付相关错误 (5000-5999) ====================
	CodePaymentNotEnabled     ErrorCode = 5000 // 支付方式未启用
	CodePaymentCreateFailed   ErrorCode = 5001 // 创建支付失败
	CodePaymentCaptureFailed  ErrorCode = 5002 // 捕获支付失败
	CodePaymentVerifyFailed   ErrorCode = 5003 // 支付验证失败
	CodePaymentNotCompleted   ErrorCode = 5004 // 支付未完成
	CodePaymentMethodInvalid  ErrorCode = 5005 // 支付方式无效
	CodeBalanceNotEnough      ErrorCode = 5006 // 余额不足
	CodeRechargeOrderNotFound ErrorCode = 5007 // 充值订单不存在
	CodeCouponNotFound        ErrorCode = 5008 // 优惠券不存在
	CodeCouponInvalid         ErrorCode = 5009 // 优惠券无效
	CodeCouponExpired         ErrorCode = 5010 // 优惠券已过期
	CodeCouponUsed            ErrorCode = 5011 // 优惠券已使用
	CodeCouponNotApplicable   ErrorCode = 5012 // 优惠券不适用

	// ==================== 客服系统错误 (6000-6999) ====================
	CodeTicketNotFound        ErrorCode = 6000 // 工单不存在
	CodeTicketClosed          ErrorCode = 6001 // 工单已关闭
	CodeTicketNotOwner        ErrorCode = 6002 // 不是工单所有者
	CodeChatNotFound          ErrorCode = 6003 // 聊天会话不存在
	CodeChatEnded             ErrorCode = 6004 // 聊天已结束
	CodeStaffNotFound         ErrorCode = 6005 // 客服不存在
	CodeStaffOffline          ErrorCode = 6006 // 客服不在线
	CodeStaffLoginFailed      ErrorCode = 6007 // 客服登录失败
	CodeFAQNotFound           ErrorCode = 6008 // FAQ不存在
	CodeKnowledgeNotFound     ErrorCode = 6009 // 知识库文章不存在
	CodeTemplateNotFound      ErrorCode = 6010 // 工单模板不存在

	// ==================== 管理后台错误 (7000-7999) ====================
	CodeAdminNotFound       ErrorCode = 7000 // 管理员不存在
	CodeAdminExists         ErrorCode = 7001 // 管理员已存在
	CodeAdminLoginFailed    ErrorCode = 7002 // 管理员登录失败
	CodeAdminDisabled       ErrorCode = 7003 // 管理员已禁用
	CodeAdminNoPermission   ErrorCode = 7004 // 无操作权限
	CodeRoleNotFound        ErrorCode = 7005 // 角色不存在
	CodeRoleInUse           ErrorCode = 7006 // 角色正在使用
	CodeBackupFailed        ErrorCode = 7007 // 备份失败
	CodeBackupNotFound      ErrorCode = 7008 // 备份不存在
	CodeRestoreFailed       ErrorCode = 7009 // 恢复失败
	CodeExportFailed        ErrorCode = 7010 // 导出失败
	CodeUndoNotFound        ErrorCode = 7011 // 撤销操作不存在
	CodeUndoExpired         ErrorCode = 7012 // 撤销操作已过期
	CodeTaskNotFound        ErrorCode = 7013 // 任务不存在
	CodeTaskRunFailed       ErrorCode = 7014 // 任务执行失败

	// ==================== 系统配置错误 (8000-8999) ====================
	CodeConfigNotFound     ErrorCode = 8000 // 配置不存在
	CodeConfigSaveFailed   ErrorCode = 8001 // 配置保存失败
	CodeDBConfigError      ErrorCode = 8004 // 数据库配置错误
	CodeEmailConfigError   ErrorCode = 8005 // 邮箱配置错误
	CodePaymentConfigError ErrorCode = 8006 // 支付配置错误
)

// errorMessages 错误码对应的错误消息
var errorMessages = map[ErrorCode]string{
	// 成功
	CodeSuccess: "操作成功",

	// 通用错误
	CodeInternalError:      "服务器内部错误",
	CodeParamError:         "参数错误",
	CodeUnauthorized:       "请先登录",
	CodeForbidden:          "无权访问",
	CodeNotFound:           "资源不存在",
	CodeTooManyRequests:    "请求过于频繁，请稍后再试",
	CodeServiceUnavailable: "服务暂不可用",
	CodeDatabaseError:      "数据库错误",
	CodeValidationError:    "数据验证失败",
	CodeConflict:           "资源已存在",
	CodeCSRFError:          "CSRF验证失败，请刷新页面重试",
	CodeCaptchaError:       "验证码错误",
	CodeServiceNotInit:     "服务未初始化",
	CodeDBNotConnected:     "数据库未连接",

	// 用户相关
	CodeUserNotFound:           "用户不存在",
	CodeUserExists:             "用户名已存在",
	CodePasswordWrong:          "密码错误",
	CodePasswordTooShort:       "密码长度至少6位",
	CodePasswordNotMatch:       "两次密码不一致",
	CodeUserDisabled:           "用户已被禁用",
	CodeLoginFailed:            "登录失败",
	CodeLoginLocked:            "登录尝试次数过多，请稍后再试",
	CodeSessionExpired:         "登录已过期，请重新登录",
	Code2FARequired:            "请完成两步验证",
	Code2FACodeWrong:           "验证码错误",
	CodeEmailNotVerified:       "请先验证邮箱",
	CodeEmailVerifyFailed:      "邮箱验证失败",
	CodeEmailCodeWrong:         "邮箱验证码错误",
	CodeEmailCodeExpired:       "邮箱验证码已过期",
	CodeEmailExists:            "该邮箱已被绑定",
	CodeEmailSendFailed:        "邮件发送失败",
	CodeResetTokenInvalid:      "重置令牌无效",
	CodeResetTokenExpired:      "重置令牌已过期，请重新申请",
	CodeUsernameInvalid:        "用户名格式无效",
	CodePhoneInvalid:           "手机号格式无效",
	CodeDeviceNotFound:         "设备不存在",
	CodeAccountDeletionPending: "账户注销申请处理中",
	CodeTOTPNotSet:             "未设置动态口令",
	CodeTOTPCodeWrong:          "动态口令错误",

	// 订单相关
	CodeOrderNotFound:      "订单不存在",
	CodeOrderStatusError:   "订单状态异常",
	CodeOrderExpired:       "订单已过期",
	CodeOrderCanceled:      "订单已取消",
	CodeOrderPaid:          "订单已支付",
	CodeOrderCreateFailed:  "订单创建失败",
	CodeOrderNotOwner:      "无权操作此订单",
	CodeOrderCannotCancel:  "订单无法取消",
	CodeKamiNotAvailable:   "卡密暂无库存",
	CodeKamiGetFailed:      "获取卡密失败",
	CodeInvoiceNotFound:    "发票不存在",
	CodeInvoiceApplyFailed: "发票申请失败",

	// 商品相关
	CodeProductNotFound:   "商品不存在",
	CodeProductOffline:    "商品已下架",
	CodeProductSoldOut:    "商品已售罄",
	CodeCategoryNotFound:  "分类不存在",
	CodeProductTestDeny:   "该商品不支持测试购买",
	CodeImageUploadFailed: "图片上传失败",
	CodeImageNotFound:     "图片不存在",
	CodeReviewNotFound:    "评价不存在",
	CodeReviewExists:      "您已经评价过此商品",
	CodeReviewNotAllowed:  "您无权评价此商品",

	// 支付相关
	CodePaymentNotEnabled:     "该支付方式未启用",
	CodePaymentCreateFailed:   "创建支付订单失败",
	CodePaymentCaptureFailed:  "支付捕获失败",
	CodePaymentVerifyFailed:   "支付验证失败",
	CodePaymentNotCompleted:   "支付未完成",
	CodePaymentMethodInvalid:  "不支持的支付方式",
	CodeBalanceNotEnough:      "余额不足",
	CodeRechargeOrderNotFound: "充值订单不存在",
	CodeCouponNotFound:        "优惠券不存在",
	CodeCouponInvalid:         "优惠券无效",
	CodeCouponExpired:         "优惠券已过期",
	CodeCouponUsed:            "优惠券已使用",
	CodeCouponNotApplicable:   "优惠券不适用于此商品",

	// 客服系统
	CodeTicketNotFound:    "工单不存在",
	CodeTicketClosed:      "工单已关闭",
	CodeTicketNotOwner:    "无权操作此工单",
	CodeChatNotFound:      "聊天会话不存在",
	CodeChatEnded:         "聊天已结束",
	CodeStaffNotFound:     "客服不存在",
	CodeStaffOffline:      "暂无客服在线",
	CodeStaffLoginFailed:  "客服登录失败",
	CodeFAQNotFound:       "常见问题不存在",
	CodeKnowledgeNotFound: "知识库文章不存在",
	CodeTemplateNotFound:  "工单模板不存在",

	// 管理后台
	CodeAdminNotFound:     "管理员不存在",
	CodeAdminExists:       "管理员用户名已存在",
	CodeAdminLoginFailed:  "管理员登录失败",
	CodeAdminDisabled:     "管理员账户已禁用",
	CodeAdminNoPermission: "您没有执行此操作的权限",
	CodeRoleNotFound:      "角色不存在",
	CodeRoleInUse:         "该角色正在使用中，无法删除",
	CodeBackupFailed:      "数据库备份失败",
	CodeBackupNotFound:    "备份文件不存在",
	CodeRestoreFailed:     "数据恢复失败",
	CodeExportFailed:      "数据导出失败",
	CodeUndoNotFound:      "撤销操作不存在",
	CodeUndoExpired:       "撤销操作已过期",
	CodeTaskNotFound:      "定时任务不存在",
	CodeTaskRunFailed:     "任务执行失败",

	// 系统配置
	CodeConfigNotFound:     "配置不存在",
	CodeConfigSaveFailed:   "配置保存失败",
	CodeDBConfigError:      "数据库配置错误",
	CodeEmailConfigError:   "邮箱配置错误",
	CodePaymentConfigError: "支付配置错误",
}

// GetMessage 获取错误码对应的消息
func (code ErrorCode) GetMessage() string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// ToInt 转换为 int
func (code ErrorCode) ToInt() int {
	return int(code)
}

// IsSuccess 是否为成功
func (code ErrorCode) IsSuccess() bool {
	return code == CodeSuccess
}

// ==================== 业务错误结构 ====================

// BusinessError 业务错误
type BusinessError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}

// NewBusinessError 创建业务错误
func NewBusinessError(code ErrorCode) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: code.GetMessage(),
	}
}

// NewBusinessErrorWithDetail 创建带详细信息的业务错误
func NewBusinessErrorWithDetail(code ErrorCode, detail string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: code.GetMessage(),
		Detail:  detail,
	}
}

// NewBusinessErrorWithMessage 创建自定义消息的业务错误
func NewBusinessErrorWithMessage(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// ==================== HTTP 状态码映射 ====================

// codeToHTTPStatus 错误码到 HTTP 状态码的映射
var codeToHTTPStatus = map[ErrorCode]int{
	// 成功
	CodeSuccess: 200,

	// 通用错误
	CodeParamError:         400,
	CodeUnauthorized:       401,
	CodeForbidden:          403,
	CodeNotFound:           404,
	CodeTooManyRequests:    429,
	CodeInternalError:      500,
	CodeServiceUnavailable: 503,
	CodeDatabaseError:      500,
	CodeValidationError:    400,
	CodeConflict:           409,
	CodeCSRFError:          403,
	CodeCaptchaError:       400,
	CodeServiceNotInit:     503,
	CodeDBNotConnected:     503,

	// 用户相关
	CodeUserNotFound:      404,
	CodeUserExists:        409,
	CodePasswordWrong:     401,
	CodePasswordTooShort:  400,
	CodePasswordNotMatch:  400,
	CodeUserDisabled:      403,
	CodeLoginFailed:       401,
	CodeLoginLocked:       429,
	CodeSessionExpired:    401,
	Code2FARequired:       401,
	Code2FACodeWrong:      401,
	CodeEmailNotVerified:  403,
	CodeEmailCodeWrong:    400,
	CodeEmailCodeExpired:  400,
	CodeEmailSendFailed:   500,
	CodeResetTokenInvalid: 400,

	// 订单相关
	CodeOrderNotFound:    404,
	CodeOrderStatusError: 400,
	CodeOrderNotOwner:    403,
	CodeKamiNotAvailable: 503,

	// 商品相关
	CodeProductNotFound: 404,
	CodeProductOffline:  400,
	CodeProductSoldOut:  400,
	CodeProductTestDeny: 400,

	// 支付相关
	CodePaymentNotEnabled:    400,
	CodePaymentCreateFailed:  500,
	CodePaymentVerifyFailed:  400,
	CodeBalanceNotEnough:     400,
	CodeCouponNotFound:       404,
	CodeCouponInvalid:        400,
	CodeCouponExpired:        400,

	// 客服系统
	CodeTicketNotFound: 404,
	CodeTicketClosed:   400,
	CodeTicketNotOwner: 403,
	CodeChatNotFound:   404,
	CodeStaffOffline:   503,

	// 管理后台
	CodeAdminNotFound:     404,
	CodeAdminNoPermission: 403,
	CodeBackupFailed:      500,
	CodeExportFailed:      500,
}

// GetHTTPStatus 获取错误码对应的 HTTP 状态码
func (code ErrorCode) GetHTTPStatus() int {
	if status, ok := codeToHTTPStatus[code]; ok {
		return status
	}
	// 默认根据错误码范围推断
	switch {
	case code >= 1000 && code < 2000:
		return 500
	case code >= 2000 && code < 3000:
		return 400
	case code >= 3000 && code < 4000:
		return 400
	case code >= 4000 && code < 5000:
		return 400
	case code >= 5000 && code < 6000:
		return 400
	case code >= 6000 && code < 7000:
		return 400
	case code >= 7000 && code < 8000:
		return 400
	case code >= 8000 && code < 9000:
		return 500
	default:
		return 500
	}
}

// ==================== Gin 响应扩展 ====================

// RespondWithCode 使用错误码响应
// 参数：
//   - c: Gin 上下文
//   - code: 错误码
//
// 响应格式：{"success": false, "code": code, "error": message}
func RespondWithCode(c *gin.Context, code ErrorCode) {
	c.JSON(code.GetHTTPStatus(), gin.H{
		"success": code.IsSuccess(),
		"code":    code.ToInt(),
		"error":   code.GetMessage(),
	})
}

// RespondWithCodeAndDetail 使用错误码和详细信息响应
func RespondWithCodeAndDetail(c *gin.Context, code ErrorCode, detail string) {
	message := code.GetMessage()
	if detail != "" {
		message = message + ": " + detail
	}
	c.JSON(code.GetHTTPStatus(), gin.H{
		"success": code.IsSuccess(),
		"code":    code.ToInt(),
		"error":   message,
	})
}

// RespondWithBusinessError 使用业务错误响应
func RespondWithBusinessError(c *gin.Context, err *BusinessError) {
	message := err.Message
	if err.Detail != "" {
		message = message + ": " + err.Detail
	}
	c.JSON(err.Code.GetHTTPStatus(), gin.H{
		"success": false,
		"code":    err.Code.ToInt(),
		"error":   message,
	})
}
