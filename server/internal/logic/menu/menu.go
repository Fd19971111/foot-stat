package menu

import v1 "server/api/route/v1"

const Home = "home"

// 常量路由与前端 src/router/elegant/routes.ts 中 constant: true 的路由保持一致
func GetConstantRoutes() []*v1.MenuRoute {
	return []*v1.MenuRoute{
		{
			Id: "403", Name: "403", Path: "/403", Component: "layout.blank$view.403",
			Meta: v1.RouteMeta{Title: "403", I18nKey: "route.403", Constant: true, HideInMenu: true},
		},
		{
			Id: "404", Name: "404", Path: "/404", Component: "layout.blank$view.404",
			Meta: v1.RouteMeta{Title: "404", I18nKey: "route.404", Constant: true, HideInMenu: true},
		},
		{
			Id: "500", Name: "500", Path: "/500", Component: "layout.blank$view.500",
			Meta: v1.RouteMeta{Title: "500", I18nKey: "route.500", Constant: true, HideInMenu: true},
		},
		{
			Id: "iframe-page", Name: "iframe-page", Path: "/iframe-page/:url", Component: "layout.base$view.iframe-page", Props: true,
			Meta: v1.RouteMeta{Title: "iframe-page", I18nKey: "route.iframe-page", Constant: true, HideInMenu: true, KeepAlive: true},
		},
		{
			Id: "login", Name: "login", Path: "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			Component: "layout.blank$view.login", Props: true,
			Meta: v1.RouteMeta{Title: "login", I18nKey: "route.login", Constant: true, HideInMenu: true},
		},
	}
}

// GetUserRoutes 按角色下发菜单，后续可替换为 sys_menu + 角色授权表查询
func GetUserRoutes(roles []string) []*v1.MenuRoute {
	routes := []*v1.MenuRoute{
		{
			Id: "home", Name: "home", Path: "/home", Component: "layout.base$view.home",
			Meta: v1.RouteMeta{Title: "比赛中心", Icon: "mdi:view-dashboard-outline", Order: 1},
		},
		{
			Id: "import", Name: "import", Path: "/import", Component: "layout.base$view.import",
			Meta: v1.RouteMeta{Title: "比赛导入", Icon: "mdi:tray-arrow-up", Order: 2},
		},
		{
			Id: "analysis", Name: "analysis", Path: "/analysis", Component: "layout.base$view.analysis",
			Meta: v1.RouteMeta{Title: "分析结果", Icon: "mdi:chart-box-outline", Order: 3},
		},
		{
			Id: "results", Name: "results", Path: "/results", Component: "layout.base$view.results",
			Meta: v1.RouteMeta{Title: "赛果记录", Icon: "mdi:clipboard-text-outline", Order: 4},
		},
		{
			Id: "history", Name: "history", Path: "/history", Component: "layout.base$view.history",
			Meta: v1.RouteMeta{Title: "数据分析", Icon: "mdi:chart-line", Order: 5},
		},
	}

	return routes
}

// IsRouteExist 判断路由名是否属于常量路由或当前可下发的菜单
func IsRouteExist(name string) bool {
	for _, route := range GetConstantRoutes() {
		if route.Name == name {
			return true
		}
	}
	for _, route := range GetUserRoutes([]string{"R_SUPER"}) {
		if route.Name == name || hasChild(route, name) {
			return true
		}
	}
	return false
}

func hasChild(route *v1.MenuRoute, name string) bool {
	for _, child := range route.Children {
		if child.Name == name || hasChild(child, name) {
			return true
		}
	}
	return false
}
