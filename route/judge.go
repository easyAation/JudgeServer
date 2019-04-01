package route

import (
	"net/http"

	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"

	"github.com/gin-gonic/gin"
)

func JudgeRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter("v1/judge_problem", http.MethodPost, reply.Wrap(judgeProblem)),
	}

	return router.ModuleRoute{
		Routers: routes,
	}
}

func judgeProblem(ctx *gin.Context) gin.HandlerFunc {

	return reply.Success(http.StatusOK, nil)
}
