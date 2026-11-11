package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"server/internal/controller/auth"
	footballController "server/internal/controller/football"
	"server/internal/controller/hello"
	routeController "server/internal/controller/route"
	"server/internal/logic/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareCORS)
				group.Middleware(ghttp.MiddlewareHandlerResponse)

				// 免鉴权：登录、刷新令牌、连通性探测
				group.Bind(
					auth.NewV1(),
					hello.NewV1(),
					footballController.NewV1(),
				)

				// 需鉴权：用户信息、动态菜单
				group.Group("/", func(private *ghttp.RouterGroup) {
					private.Middleware(middleware.Auth)
					private.Bind(
						auth.NewV1Plus(),
						routeController.NewV1(),
					)
				})
			})
			s.Run()
			return nil
		},
	}
)
