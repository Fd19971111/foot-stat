package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"server/api/auth"
	v1 "server/api/auth/v1"
	"server/internal/consts"
	authLogic "server/internal/logic/auth"
)

// ControllerV1 免鉴权：登录、刷新令牌
type ControllerV1 struct{}

func NewV1() auth.IAuthV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error) {
	return authLogic.Login(ctx, req.UserName, req.Password)
}

func (c *ControllerV1) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (*v1.RefreshTokenRes, error) {
	return authLogic.RefreshToken(ctx, req.RefreshToken)
}

// ControllerV1Plus 需鉴权：由中间件注入的用户信息
type ControllerV1Plus struct{}

func NewV1Plus() auth.IAuthV1Plus {
	return &ControllerV1Plus{}
}

func (c *ControllerV1Plus) GetUserInfo(ctx context.Context, _ *v1.GetUserInfoReq) (*v1.GetUserInfoRes, error) {
	info, ok := ctx.Value(consts.CtxUserInfoKey).(*v1.GetUserInfoRes)
	if !ok {
		return nil, gerror.NewCode(gcode.New(consts.CodeNotLogin, "", nil), "请先登录")
	}
	return info, nil
}
