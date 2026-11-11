package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"server/internal/consts"
	authLogic "server/internal/logic/auth"
)

// Auth 校验 Bearer 令牌并把用户信息写入 context
func Auth(r *ghttp.Request) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		abort(r, consts.CodeNotLogin, "请先登录")
		return
	}

	info, err := authLogic.ValidateToken(r.Context(), token)
	if err != nil {
		abort(r, gerror.Code(err).Code(), err.Error())
		return
	}

	r.SetCtxVar(consts.CtxUserIdKey, info.UserID)
	r.SetCtxVar(consts.CtxUserNameKey, info.UserName)
	r.SetCtxVar(consts.CtxRolesKey, info.Roles)
	r.SetCtxVar(consts.CtxUserInfoKey, info)
	r.Middleware.Next()
}

// abort 只写入错误，不再调用 Middleware.Next，由外层 MiddlewareHandlerResponse 统一输出 {code,message,data}
func abort(r *ghttp.Request, code int, message string) {
	r.SetError(gerror.NewCode(gcode.New(code, "", nil), message))
}
