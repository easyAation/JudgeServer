package main

import (
	"flag"
	"fmt"
	"github.com/easyAation/scaffold/db"
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
}

func main() {
	engine := router.BuildHandler(nil, route.JudgeRouteModule())
	engine.Run(":" + strconv.Itoa(common.Config.Listen))
}
