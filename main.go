package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/easyAation/scaffold/db"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/route"
)

var (
	configPath = flag.String("conf", "conf/config.dev.toml", "config file path.")
)

func init() {
	flag.Parse()

	common.InitConfig(*configPath)

	// if !route.Exists(common.Config.Compile.CodeDir) {
	// 	if err := os.MkdirAll(common.Config.Compile.CodeDir, os.ModePerm); err != nil {
	// 		panic(err)
	// 	}
	// }
	//
	// if !route.Exists(common.Config.Compile.ExeDir) {
	// 	if err := os.MkdirAll(common.Config.Compile.ExeDir, os.ModePerm); err != nil {
	// 		panic(err)
	// 	}
	// }
	fmt.Println(common.Config.MySQL)
	if err := db.RigisterDB("problem", &common.Config.MySQL); err != nil {
		panic(err)
	}

	if err := db.InitRedis(common.Config.Redis); err != nil {
		panic(err)
	}
}

func main() {
	engine := router.BuildHandler(optionsHandle, []router.MiddleWare{Cors}, route.JudgeRouteModule(),
		route.AccountRouteModule(), route.ResourceRouteModule())
	if err := engine.Run(":" + strconv.Itoa(common.Config.Listen)); err != nil {
		panic(err)
	}
}
func optionsHandle(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.AbortWithStatus(http.StatusNoContent)
}

// 处理跨域请求,支持options访问
func Cors(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		fmt.Println(method)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 放行所有OPTIONS方法，因为有的模板是要请求两次的
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		fn(c)
	}
}
