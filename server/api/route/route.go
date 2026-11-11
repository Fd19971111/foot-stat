package route

import (
	"context"

	"server/api/route/v1"
)

// IRouteV1 动态路由接口，均需鉴权
type IRouteV1 interface {
	GetUserRoutes(ctx context.Context, req *v1.GetUserRoutesReq) (res *v1.GetUserRoutesRes, err error)
	GetConstantRoutes(ctx context.Context, req *v1.GetConstantRoutesReq) (res v1.GetConstantRoutesRes, err error)
	IsRouteExist(ctx context.Context, req *v1.IsRouteExistReq) (res v1.IsRouteExistRes, err error)
}
