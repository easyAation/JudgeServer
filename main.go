package main

import (
	"flag"
	"strconv"

	"github.com/easyAation/scaffold/router"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/route"
)

var (
	configPath = flag.String("conf", "conf/config.toml", "config file path.")
)

func init() {
	flag.Parse()

	common.InitConfig(*configPath)
}

func main() {
	engine := router.BuildHandler(nil, route.JudgeRouteModule())
	engine.Run(":" + strconv.Itoa(common.Config.Listen))
}
