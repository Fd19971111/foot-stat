package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"Auth" summary:"站点用户登录"`
	UserName string `json:"userName" v:"required" dc:"登录账号"`
	Password string `json:"password" v:"required" dc:"登录密码"`
	Client   string `json:"client" dc:"客户端类型"`
}

type LoginRes struct {
	UserId       string `json:"userId"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenReq struct {
	g.Meta       `path:"/auth/refreshToken" method:"post" tags:"Auth" summary:"刷新令牌"`
	RefreshToken string `json:"refreshToken" v:"required" dc:"登录时返回的refreshToken"`
}

type RefreshTokenRes struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type GetUserInfoReq struct {
	g.Meta `path:"/auth/getUserInfo" method:"get" tags:"Auth" summary:"获取当前登录用户信息"`
}

type GetUserInfoRes struct {
	UserID   string   `json:"userId"`
	UserName string   `json:"userName"`
	Roles    []string `json:"roles"`
	Buttons  []string `json:"buttons"`
}
