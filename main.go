package main

import (
	"flag"
	"strconv"

	"scaffold/criteria/router"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/judge"
)

var (
	configPath = flag.String("conf", "conf/config.toml", "config file path.")
)

func init() {
	flag.Parse()

	common.InitConfig(*configPath)
}

func main() {
	engine := router.BuildHandler(nil, judge.JudgeRouteModule())
	engine.Run(":" + strconv.Itoa(common.Config.Listen))
}
