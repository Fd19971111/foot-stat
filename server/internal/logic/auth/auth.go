package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "server/api/auth/v1"
	"server/internal/consts"
)

type account struct {
	userName string
	password string
	userInfo v1.GetUserInfoRes
}

// 演示账号，后续可替换为数据库用户表查询（参考 yiminjia 的 site_user + bcrypt）
var accounts = []account{
	{userName: "admin", password: "123456", userInfo: v1.GetUserInfoRes{UserID: "1", UserName: "admin", Roles: []string{"R_SUPER"}, Buttons: []string{"*"}}},
	{userName: "Soybean", password: "123456", userInfo: v1.GetUserInfoRes{UserID: "1", UserName: "Soybean", Roles: []string{"R_SUPER"}, Buttons: []string{"*"}}},
	{userName: "Super", password: "123456", userInfo: v1.GetUserInfoRes{UserID: "1", UserName: "Super", Roles: []string{"R_SUPER"}, Buttons: []string{"*"}}},
	{userName: "Admin", password: "123456", userInfo: v1.GetUserInfoRes{UserID: "2", UserName: "Admin", Roles: []string{"R_ADMIN"}, Buttons: []string{}}},
	{userName: "User", password: "123456", userInfo: v1.GetUserInfoRes{UserID: "3", UserName: "User", Roles: []string{"R_USER"}, Buttons: []string{}}},
}

type session struct {
	userInfo       v1.GetUserInfoRes
	token          string
	refreshToken   string
	expireAt       time.Time
	refreshExpires time.Time
}

type store struct {
	mu        sync.RWMutex
	byToken   map[string]*session
	byRefresh map[string]*session
}

var sessions = &store{
	byToken:   make(map[string]*session),
	byRefresh: make(map[string]*session),
}

func findAccount(userName string) (account, bool) {
	for _, item := range accounts {
		if strings.EqualFold(item.userName, userName) {
			return item, true
		}
	}
	return account{}, false
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func durationConfig(ctx context.Context, key string, def time.Duration) time.Duration {
	raw := g.Cfg().MustGet(ctx, key).String()
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func (s *store) remove(sess *session) {
	delete(s.byToken, sess.token)
	delete(s.byRefresh, sess.refreshToken)
}

func notLoginError(message string) error {
	return gerror.NewCode(gcode.New(consts.CodeNotLogin, "", nil), message)
}

func expiredError(message string) error {
	return gerror.NewCode(gcode.New(consts.CodeTokenExpired, "", nil), message)
}

// Login 校验账号并签发访问令牌与刷新令牌，同一用户的新登录会踢掉旧会话
func Login(ctx context.Context, userName, password string) (*v1.LoginRes, error) {
	acc, ok := findAccount(userName)
	if !ok || acc.password != password {
		return nil, gerror.New("用户名或密码错误")
	}

	now := time.Now()
	sess := &session{
		userInfo:       acc.userInfo,
		token:          randomToken(),
		refreshToken:   randomToken(),
		expireAt:       now.Add(durationConfig(ctx, "app.tokenExpire", consts.DefaultTokenExpire)),
		refreshExpires: now.Add(durationConfig(ctx, "app.refreshTokenExpire", consts.DefaultRefreshTokenExpire)),
	}

	sessions.mu.Lock()
	for _, old := range sessions.byToken {
		if old.userInfo.UserID == acc.userInfo.UserID {
			sessions.remove(old)
		}
	}
	sessions.byToken[sess.token] = sess
	sessions.byRefresh[sess.refreshToken] = sess
	sessions.mu.Unlock()

	return &v1.LoginRes{UserId: acc.userInfo.UserID, Token: sess.token, RefreshToken: sess.refreshToken}, nil
}

// RefreshToken 轮换令牌并滑动续期；刷新令牌失效返回登出码，避免前端陷入刷新死循环
func RefreshToken(ctx context.Context, refreshToken string) (*v1.RefreshTokenRes, error) {
	sessions.mu.Lock()
	defer sessions.mu.Unlock()

	sess, ok := sessions.byRefresh[refreshToken]
	if !ok {
		return nil, notLoginError("请先登录")
	}
	if time.Now().After(sess.refreshExpires) {
		sessions.remove(sess)
		return nil, notLoginError("登录已过期，请重新登录")
	}

	sessions.remove(sess)
	sess.token = randomToken()
	sess.refreshToken = randomToken()
	sess.expireAt = time.Now().Add(durationConfig(ctx, "app.tokenExpire", consts.DefaultTokenExpire))
	sess.refreshExpires = time.Now().Add(durationConfig(ctx, "app.refreshTokenExpire", consts.DefaultRefreshTokenExpire))
	sessions.byToken[sess.token] = sess
	sessions.byRefresh[sess.refreshToken] = sess

	return &v1.RefreshTokenRes{Token: sess.token, RefreshToken: sess.refreshToken}, nil
}

// ValidateToken 校验访问令牌，返回用户信息；无效返回 8888，过期返回 9999
func ValidateToken(_ context.Context, token string) (*v1.GetUserInfoRes, error) {
	sessions.mu.RLock()
	sess, ok := sessions.byToken[token]
	sessions.mu.RUnlock()
	if !ok {
		return nil, notLoginError("请先登录")
	}
	if time.Now().After(sess.expireAt) {
		return nil, expiredError("登录已过期，请重新登录")
	}

	info := sess.userInfo
	return &info, nil
}
