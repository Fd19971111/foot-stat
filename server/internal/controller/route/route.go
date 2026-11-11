package route

import (
	"context"

	"server/api/route"
	v1 "server/api/route/v1"
	"server/internal/consts"
	menuLogic "server/internal/logic/menu"
)

type ControllerV1 struct{}

func NewV1() route.IRouteV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) GetUserRoutes(ctx context.Context, _ *v1.GetUserRoutesReq) (*v1.GetUserRoutesRes, error) {
	roles, _ := ctx.Value(consts.CtxRolesKey).([]string)

	return &v1.GetUserRoutesRes{
		UserRoute: v1.UserRoute{Routes: menuLogic.GetUserRoutes(roles), Home: menuLogic.Home},
	}, nil
}

func (c *ControllerV1) GetConstantRoutes(_ context.Context, _ *v1.GetConstantRoutesReq) (v1.GetConstantRoutesRes, error) {
	return menuLogic.GetConstantRoutes(), nil
}

func (c *ControllerV1) IsRouteExist(_ context.Context, req *v1.IsRouteExistReq) (v1.IsRouteExistRes, error) {
	return v1.IsRouteExistRes(menuLogic.IsRouteExist(req.RouteName)), nil
}
