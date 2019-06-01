package route

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/utils"
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
		router.NewRouter(
			"/v1/userRank",
			http.MethodGet,
			reply.Wrap(userRank),
		),
		
	}
	return router.ModuleRoute{
		Routers: routes,
	}
}

//
func userRank(ctx *gin.Context) gin.HandlerFunc {
	allSubmits, err := model.GetSubmits(ctx, nil)
	if err != nil {
		return reply.Err(err)
	}

	var (
		submitCount = make(map[string]int)
		acceptCount = make(map[string]int)
		already     = make(map[string]bool)
	)

	for _, val := range allSubmits {
		submitCount[val.UID]++
		if val.Result == common.Accept {
			if _, exist := already[val.UID+fmt.Sprintf("%d", val.PID)]; !exist {
				acceptCount[val.UID]++
			}
			already[val.UID+fmt.Sprintf("%d", val.PID)] = true
		}
	}

	allAccount, err := model.GetAccounts(ctx, nil)
	if err != nil {
		return reply.Err(err)
	}

	type user struct {
		model.Account
		Submit int `json:"submit"`
		Accept int `json:"accept"`
	}

	userInfo := make([]user, 0)
	for _, ac := range allAccount {
		userInfo = append(userInfo, user{
			ac,
			submitCount[ac.ID],
			acceptCount[ac.ID],
		})
	}

	sort.Slice(userInfo, func(i, j int) bool {
		if userInfo[i].Accept != userInfo[j].Accept {
			return userInfo[i].Accept > userInfo[j].Accept
		}
		return userInfo[i].Submit < userInfo[j].Submit
	})

	return reply.Success(200, map[string]interface{}{
		"total": len(allAccount),
		"data":  userInfo,
	})
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
