package judge

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"scaffold/criteria/reply"
	"scaffold/criteria/router"
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
