package route

import (
	"net/http"
	"online_judge/JudgeServer/utils"

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
	p := struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Password   string `json:"password"`
		GithupAddr string `json:"githup_addr"`
		BlogAddr   string `json:"blog_addr"`
	}{}
	err := ctx.ShouldBind(&p)
	if err != nil {
		return reply.Err(err)
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	err = model.RegisterAccout(sqlExec, model.Account{
		ID:         p.ID,
		Name:       p.Name,
		Auth:       utils.EncryptPassword(p.ID, p.Password),
		GitHupAddr: p.GithupAddr,
		BlogAddr:   p.BlogAddr,
	})
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, nil)
}
