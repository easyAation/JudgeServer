package route

import (
	"net/http"

	"github.com/easyAation/scaffold/db"
	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"

	"online_judge/JudgeServer/model"
)

func AccountRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter(
			"v1/account/register",
			http.MethodPost,
			reply.Wrap(registerAccount),
		),
	}
	return router.ModuleRoute{
		Routers: routes,
	}
}
func registerAccount(ctx *gin.Context) gin.HandlerFunc {
	var ac model.Account
	err := ctx.ShouldBind(&ac)
	if err != nil {
		return reply.Err(err)
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	err = model.RegisterAccout(sqlExec, ac)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, nil)
}
