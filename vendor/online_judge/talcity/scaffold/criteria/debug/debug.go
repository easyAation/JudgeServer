package debug

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"

	"online_judge/talcity/scaffold/criteria/reply"
	"online_judge/talcity/scaffold/criteria/route"
)

func NewModuleRoute() *route.ModuleRoute {
	var (
		routes         []*route.Route
		ignorePrevious bool
		middles        []route.Middleware
	)
	routes = []*route.Route{
		route.NewRoute(
			"/internal/debug",
			"GET",
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, r.RequestURI+"/pprof/", 301)
			},
		),
		route.NewRoute(
			"/internal/debug/vars",
			"GET",
			expVars,
		),
		route.NewRoute(
			"/internal/debug/err_detail_swith",
			"POST",
			switchRespErrDetail,
		),
		route.NewRoute(
			"/internal/debug/pprof/", // 要保证path是/internal/debug/pprof/ 而不是/internal/debug/
			"GET",
			pprof.Index,
		),
		route.NewRoute(
			"/internal/debug/pprof/cmdline",
			"GET",
			pprof.Cmdline,
		),
		route.NewRoute(
			"/internal/debug/pprof/profile",
			"GET",
			pprof.Profile,
		),
		route.NewRoute(
			"/internal/debug/pprof/symbol",
			"GET",
			pprof.Symbol,
		),
		route.NewRoute(
			"/internal/debug/pprof/trace",
			"GET",
			pprof.Trace,
		),
		route.NewRoute(
			"/internal/debug/pprof/block",
			"GET",
			pprof.Handler("block").ServeHTTP,
		),
		route.NewRoute(
			"/internal/debug/pprof/heap",
			"GET",
			pprof.Handler("heap").ServeHTTP,
		),
		route.NewRoute(
			"/internal/debug/pprof/goroutine",
			"GET",
			pprof.Handler("goroutine").ServeHTTP,
		),
		route.NewRoute(
			"/internal/debug/pprof/threadcreate",
			"GET",
			pprof.Handler("threadcreate").ServeHTTP,
		),
	}

	ignorePrevious = true
	return &route.ModuleRoute{
		Routes: routes,
		IgnorePreviousMiddleware: ignorePrevious,
		Middlewares:              middles,
	}
}

func expVars(w http.ResponseWriter, r *http.Request) {
	first := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

// 调试开关, 打开后如果API返回错误, 增加description字段描述错误详情, 关闭后不显示description字段
func switchRespErrDetail(w http.ResponseWriter, r *http.Request) {
	swcStr := r.URL.Query().Get("swc")
	swc, err := strconv.ParseBool(swcStr)
	if err != nil {
		fmt.Fprintln(w, "failed")
		return
	}
	reply.SwitchRespErrDetail(swc)

	fmt.Fprintln(w, "OK")
	return
}
