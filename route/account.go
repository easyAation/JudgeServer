package route

import (
	"fmt"
	"net/http"
	"online_judge/JudgeServer/utils"

	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"online_judge/JudgeServer/model"
)

func AccountRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter(
			"/v1/account/register",
			http.MethodPost,
			reply.Wrap(registerAccount),
		),
		router.NewRouter(
			"/v1/signin",
			http.MethodPost,
			reply.Wrap(signin),
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
	err := ctx.ShouldBindJSON(&p)
	if err != nil {
		return reply.Err(err)
	}
	fmt.Println(p)
	err = model.RegisterAccount(ctx, model.Account{
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

func signin(ctx *gin.Context) gin.HandlerFunc {
	p := struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}{}
	err := ctx.ShouldBindJSON(&p)
	if err != nil {
		return reply.Err(err)
	}
	ac, err := model.GetOneAccount(ctx, map[string]interface{}{
		"id": p.ID,
	})
	if err != nil {
		return reply.Err(err)
	}

	if utils.EncryptPassword(ac.ID, p.Password) != ac.Auth {
		return reply.Err(errors.Errorf("invalid password"))
	}

	token, err := utils.CreateToken(ctx, p.ID, p.Password, ac.Auth)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"token": token,
		"name":  ac.Name,
	})
}
