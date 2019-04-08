package route

import (
	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"online_judge/JudgeServer/sandbox"
	"os"
)

func JudgeRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter(
			"/v1/judge_problem",
			http.MethodPost,
			reply.Wrap(judgeProblem),
		),
	}

	return router.ModuleRoute{
		Routers: routes,
	}
}

func judgeProblem(ctx *gin.Context) gin.HandlerFunc {
	var (
		request sandbox.Request
	)
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		return reply.ErrorWithMessage(errors.WithStack(err), "invalid param")
	}

	sandBox, err := sandbox.NewSandBox(request)
	if err != nil {
		return reply.Err(err)
	}
	response, err := sandBox.Run()
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
