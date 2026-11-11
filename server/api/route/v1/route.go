package v1

import "github.com/gogf/gf/v2/frame/g"

type GetUserRoutesReq struct {
	g.Meta `path:"/route/getUserRoutes" method:"get" tags:"Route" summary:"获取当前用户菜单路由"`
}

type RouteMeta struct {
	Title      string   `json:"title"`
	I18nKey    string   `json:"i18nKey,omitempty"`
	Icon       string   `json:"icon,omitempty"`
	Order      int      `json:"order,omitempty"`
	HideInMenu bool     `json:"hideInMenu,omitempty"`
	Constant   bool     `json:"constant,omitempty"`
	KeepAlive  bool     `json:"keepAlive,omitempty"`
	Roles      []string `json:"roles,omitempty"`
}

type MenuRoute struct {
	Id        string       `json:"id"`
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Component string       `json:"component,omitempty"`
	Props     bool         `json:"props,omitempty"`
	Meta      RouteMeta    `json:"meta"`
	Children  []*MenuRoute `json:"children,omitempty"`
}

type UserRoute struct {
	Routes []*MenuRoute `json:"routes"`
	Home   string       `json:"home"`
}

type GetUserRoutesRes struct {
	UserRoute
}

type GetConstantRoutesReq struct {
	g.Meta `path:"/route/getConstantRoutes" method:"get" tags:"Route" summary:"获取常量路由"`
}

type GetConstantRoutesRes []*MenuRoute

type IsRouteExistReq struct {
	g.Meta    `path:"/route/isRouteExist" method:"get" tags:"Route" summary:"判断路由是否存在"`
	RouteName string `json:"routeName"`
}

type IsRouteExistRes bool
