package api

// ==================== 用户认证 API 文档注解 ====================

// UserRegister 用户注册
// @Summary      用户注册
// @Description  新用户注册，需要邮箱验证码
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request body SwaggerUserRegisterRequest true "注册信息"
// @Success      200 {object} SwaggerSuccessResponse{data=object{id=int,username=string}} "注册成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误或用户已存在"
// @Failure      500 {object} SwaggerErrorResponse "服务器错误"
// @Router       /api/user/register [post]

// UserLogin 用户登录
// @Summary      用户登录
// @Description  用户登录，如启用2FA需要额外验证
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        request body SwaggerUserLoginRequest true "登录信息"
// @Success      200 {object} SwaggerSuccessResponse{data=object{id=int,username=string,email=string}} "登录成功"
// @Success      200 {object} SwaggerSuccessResponse{data=object{require_2fa=bool,verify_token=string}} "需要二次验证"
// @Failure      400 {object} SwaggerErrorResponse "用户名或密码错误"
// @Failure      429 {object} SwaggerErrorResponse "登录次数过多"
// @Router       /api/user/login [post]

// UserLogout 用户登出
// @Summary      用户登出
// @Description  退出当前登录状态
// @Tags         用户认证
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse "登出成功"
// @Security     CookieAuth
// @Router       /api/user/logout [post]

// UserInfo 获取用户信息
// @Summary      获取当前用户信息
// @Description  获取当前登录用户的详细信息
// @Tags         用户信息
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=SwaggerUserInfo} "用户信息"
// @Failure      400 {object} SwaggerErrorResponse "用户不存在"
// @Security     CookieAuth
// @Router       /api/user/info [get]

// UpdatePassword 修改密码
// @Summary      修改密码
// @Description  修改当前用户的登录密码
// @Tags         用户信息
// @Accept       json
// @Produce      json
// @Param        request body object{old_password=string,new_password=string} true "密码信息"
// @Success      200 {object} SwaggerSuccessResponse "修改成功"
// @Failure      400 {object} SwaggerErrorResponse "旧密码错误或新密码不符合要求"
// @Security     CookieAuth
// @Router       /api/user/password [put]

// UpdateUserInfo 更新用户信息
// @Summary      更新用户信息
// @Description  更新用户的手机号等信息
// @Tags         用户信息
// @Accept       json
// @Produce      json
// @Param        request body object{phone=string} true "用户信息"
// @Success      200 {object} SwaggerSuccessResponse "更新成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误"
// @Security     CookieAuth
// @Router       /api/user/update [put]

// ==================== 两步验证 API ====================

// Get2FAStatus 获取2FA状态
// @Summary      获取两步验证状态
// @Description  获取当前用户的两步验证配置状态
// @Tags         两步验证
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{enabled=bool,prefer_email_auth=bool,email_verified=bool,has_totp=bool}} "2FA状态"
// @Security     CookieAuth
// @Router       /api/user/2fa/status [get]

// Generate2FASecret 生成2FA密钥
// @Summary      生成TOTP密钥
// @Description  生成新的TOTP密钥用于绑定验证器
// @Tags         两步验证
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse{data=object{secret=string,url=string}} "密钥信息"
// @Failure      500 {object} SwaggerErrorResponse "生成失败"
// @Security     CookieAuth
// @Router       /api/user/2fa/generate [post]

// Enable2FA 启用TOTP两步验证
// @Summary      启用TOTP两步验证
// @Description  使用TOTP验证器启用两步验证
// @Tags         两步验证
// @Accept       json
// @Produce      json
// @Param        request body object{secret=string,code=string} true "验证信息"
// @Success      200 {object} SwaggerSuccessResponse "启用成功"
// @Failure      400 {object} SwaggerErrorResponse "验证码错误"
// @Security     CookieAuth
// @Router       /api/user/2fa/enable [post]

// Enable2FAEmail 启用邮箱两步验证
// @Summary      启用邮箱两步验证
// @Description  使用邮箱验证码方式启用两步验证
// @Tags         两步验证
// @Produce      json
// @Success      200 {object} SwaggerSuccessResponse "启用成功"
// @Failure      400 {object} SwaggerErrorResponse "邮箱未验证"
// @Security     CookieAuth
// @Router       /api/user/2fa/enable-email [post]

// Disable2FA 禁用两步验证
// @Summary      禁用两步验证
// @Description  禁用两步验证功能
// @Tags         两步验证
// @Accept       json
// @Produce      json
// @Param        request body object{totp_code=string,email_code=string} true "验证信息"
// @Success      200 {object} SwaggerSuccessResponse "禁用成功"
// @Failure      400 {object} SwaggerErrorResponse "验证失败"
// @Security     CookieAuth
// @Router       /api/user/2fa/disable [post]

// Set2FAPreference 设置2FA偏好
// @Summary      设置两步验证偏好
// @Description  设置首选的两步验证方式（邮箱或TOTP）
// @Tags         两步验证
// @Accept       json
// @Produce      json
// @Param        request body object{prefer_email_auth=bool} true "偏好设置"
// @Success      200 {object} SwaggerSuccessResponse "设置成功"
// @Security     CookieAuth
// @Router       /api/user/2fa/preference [put]

// ==================== 邮箱验证 API ====================

// SendEmailCode 发送邮箱验证码
// @Summary      发送邮箱验证码
// @Description  发送验证码到指定邮箱
// @Tags         邮箱验证
// @Accept       json
// @Produce      json
// @Param        request body object{email=string,code_type=string} true "邮箱信息（code_type: register/login/reset_password/enable_2fa）"
// @Success      200 {object} SwaggerSuccessResponse "发送成功"
// @Failure      400 {object} SwaggerErrorResponse "参数错误"
// @Failure      500 {object} SwaggerErrorResponse "发送失败"
// @Router       /api/email/send-code [post]

// VerifyEmailCode 验证邮箱验证码
// @Summary      验证邮箱验证码
// @Description  验证邮箱验证码是否正确
// @Tags         邮箱验证
// @Accept       json
// @Produce      json
// @Param        request body object{email=string,code=string,code_type=string} true "验证信息"
// @Success      200 {object} SwaggerSuccessResponse "验证成功"
// @Failure      400 {object} SwaggerErrorResponse "验证码错误"
// @Router       /api/email/verify-code [post]

// BindEmail 绑定邮箱
// @Summary      绑定邮箱
// @Description  为当前用户绑定邮箱地址
// @Tags         邮箱验证
// @Accept       json
// @Produce      json
// @Param        request body object{email=string,code=string} true "绑定信息"
// @Success      200 {object} SwaggerSuccessResponse "绑定成功"
// @Failure      400 {object} SwaggerErrorResponse "验证码错误或邮箱已被使用"
// @Security     CookieAuth
// @Router       /api/user/bind-email [post]

// ==================== 找回密码 API ====================

// ForgotPasswordCheck 检查用户
// @Summary      检查用户存在性
// @Description  检查用户是否存在并返回隐藏的邮箱信息
// @Tags         找回密码
// @Accept       json
// @Produce      json
// @Param        request body object{username=string} true "用户名"
// @Success      200 {object} SwaggerSuccessResponse{data=object{email=string,masked_email=string,has_2fa=bool}} "用户信息"
// @Failure      400 {object} SwaggerErrorResponse "用户不存在"
// @Router       /api/forgot-password/check [post]

// ForgotPasswordVerify 验证身份
// @Summary      验证身份
// @Description  通过邮箱验证码或TOTP验证用户身份
// @Tags         找回密码
// @Accept       json
// @Produce      json
// @Param        request body object{username=string,email_code=string,totp_code=string} true "验证信息"
// @Success      200 {object} SwaggerSuccessResponse{data=object{reset_token=string}} "验证成功"
// @Failure      400 {object} SwaggerErrorResponse "验证失败"
// @Router       /api/forgot-password/verify [post]

// ForgotPasswordReset 重置密码
// @Summary      重置密码
// @Description  使用重置令牌设置新密码
// @Tags         找回密码
// @Accept       json
// @Produce      json
// @Param        request body object{username=string,reset_token=string,new_password=string} true "重置信息"
// @Success      200 {object} SwaggerSuccessResponse "重置成功"
// @Failure      400 {object} SwaggerErrorResponse "令牌无效或已过期"
// @Router       /api/forgot-password/reset [post]

// ==================== 二次验证登录 API ====================

// Get2FAInfo 获取二次验证信息
// @Summary      获取二次验证信息
// @Description  获取登录时的二次验证配置信息
// @Tags         二次验证登录
// @Produce      json
// @Param        token query string true "验证令牌"
// @Success      200 {object} SwaggerSuccessResponse{data=object{username=string,masked_email=string,prefer_email=bool,has_totp=bool}} "验证信息"
// @Failure      400 {object} SwaggerErrorResponse "令牌无效或已过期"
// @Router       /api/2fa/info [get]

// Verify2FALogin 完成二次验证登录
// @Summary      完成二次验证
// @Description  提交二次验证码完成登录
// @Tags         二次验证登录
// @Accept       json
// @Produce      json
// @Param        request body object{token=string,totp_code=string,email_code=string} true "验证信息"
// @Success      200 {object} SwaggerSuccessResponse{data=object{id=int,username=string,email=string}} "登录成功"
// @Failure      400 {object} SwaggerErrorResponse "验证码错误"
// @Router       /api/2fa/verify [post]
