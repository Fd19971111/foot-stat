package auth

import (
	"context"

	"server/api/auth/v1"
)

// IAuthV1 免鉴权接口：登录、刷新令牌
type IAuthV1 interface {
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (res *v1.RefreshTokenRes, err error)
}

// IAuthV1Plus 需鉴权接口：获取用户信息
type IAuthV1Plus interface {
	GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (res *v1.GetUserInfoRes, err error)
}
