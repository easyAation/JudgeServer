package main

import (
	"context"
	"flag"
	"net/http"
	"strconv"

	"online_judge/talcity/scaffold/criteria/debug"
	"online_judge/talcity/scaffold/criteria/grace"
	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/monitor"
	"online_judge/talcity/scaffold/criteria/route"
	"online_judge/talcity/scaffold/db"
	"online_judge/talcity/scaffold/middlewares"

	"online_judge/JudgeServer/common"
)

var (
	BuildTime    = "NO BUILD TIME"
	BuildGitHash = "NO GIT HASH"
	BuildGitTag  = "NO GIT TAG"
)

var (
	configPath = flag.String("conf", "conf/config.toml", "config file path.")
)

func init() {
	flag.Parse()

	common.InitConfig(*configPath)

	if err := db.RegisterRedis(common.Config.Redis); err != nil {
		panic(err)
	}

	monitor.Init(common.Config.Monitor.NameSpace, common.Config.Monitor.Subsystem)
	monitor.SetVersion(monitor.Version{
		GitHash:   BuildGitHash,
		GitTag:    BuildGitTag,
		BuildTime: BuildTime,
	})

	if common.Config.AllowCORS {
		route.AllowCORS()
	}
}

func main() {
	defer log.Sync()
	globalMiddlewares := []route.Middleware{
		middlewares.BuildHttpLogger(common.InternalError, "[judge-server]"),
	}
	handler := route.BuildHandler(globalMiddlewares,
		// TODO add judge serve router
		// 监控和版本信息等
		monitor.NewModuleRoute(),
		// debug相关
		debug.NewModuleRoute(),
	)
	log.Debug("debug: start serving at %d", common.Config.Listen)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(common.Config.Listen),
		Handler: handler,
	}
	// register cleanup function
	grace := grace.New()
	grace.Register(func() {
		server.Shutdown(context.Background())
	})

	grace.Run(server.ListenAndServe)
}
