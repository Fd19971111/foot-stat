package consts

import "time"

// 业务错误码，与前端 VITE_SERVICE_LOGOUT_CODES / VITE_SERVICE_EXPIRED_TOKEN_CODES 对齐
const (
	CodeNotLogin     = 8888 // 登录态失效，前端执行登出
	CodeTokenExpired = 9999 // 访问令牌过期，前端调用刷新令牌接口
)

// 令牌默认有效期
const (
	DefaultTokenExpire        = 2 * time.Hour
	DefaultRefreshTokenExpire = 7 * 24 * time.Hour
)

type ctxKey string

// 鉴权中间件写入 context 的键
const (
	CtxUserIdKey   ctxKey = "userId"
	CtxUserNameKey ctxKey = "userName"
	CtxRolesKey    ctxKey = "roles"
	CtxUserInfoKey ctxKey = "userInfo"
)
