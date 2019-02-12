package route

import (
	"fmt"
	"net/http"

	"github.com/beego/mux"
	"github.com/rs/cors"

	"online_judge/talcity/scaffold/criteria/log"
)

var (
	allowCORS bool
)

func AllowCORS() {
	allowCORS = true
}

type Middleware func(http.Handler) http.Handler

type Route struct {
	Pattern                  string
	Method                   string
	Handler                  http.HandlerFunc
	MiddleWares              []Middleware // 按从左到右的顺序生效
	IgnorePreviousMiddleware bool         // 若为true, 则之前的middleware失效
}

type ModuleRoute struct {
	Routes                   []*Route
	IgnorePreviousMiddleware bool
	Middlewares              []Middleware
}

func NewRoute(pattern, method string, handler http.HandlerFunc, middles ...Middleware) *Route {
	return &Route{
		Pattern:     pattern,
		Method:      method,
		Handler:     handler,
		MiddleWares: middles,
	}
}

func BuildHandler(topLevelMiddles []Middleware, moduleRoutes ...*ModuleRoute) http.Handler {
	router := mux.New()
	router.DefaultHandler(defaultHandler)

	for _, module := range moduleRoutes {
		for _, rou := range module.Routes {
			middlewareChain := chainMiddlewares(
				rou.IgnorePreviousMiddleware,
				rou.MiddleWares,
				module.IgnorePreviousMiddleware,
				module.Middlewares,
				topLevelMiddles,
			)
			var handler http.Handler = rou.Handler
			middleLength := len(middlewareChain)
			if middleLength > 0 {
				for i := 0; i < middleLength; i++ {
					wrapper := middlewareChain[middleLength-1-i]
					if wrapper != nil {
						handler = wrapper(handler)
					}
				}
			}
			router.Handle(rou.Method, rou.Pattern, handler.ServeHTTP)
		}
	}
	if allowCORS {
		handler := cors.AllowAll().Handler(router)
		return logRequest(handler)
	}
	return logRequest(router)
}

// TODO: 检查middleware是否有重复, 有则panic
func chainMiddlewares(inlineIgnore bool, inlineMiddles []Middleware, moduleIgnore bool, moduleMiddles []Middleware, globalMiddles []Middleware) []Middleware {
	if inlineIgnore {
		return inlineMiddles
	}
	var result []Middleware
	if moduleIgnore {
		result = make([]Middleware, 0, len(moduleMiddles)+len(inlineMiddles))
		result = append(result, moduleMiddles...)
		result = append(result, inlineMiddles...)
		return result
	}
	result = make([]Middleware, 0, len(globalMiddles)+len(moduleMiddles)+len(inlineMiddles))
	result = append(result, globalMiddles...)
	result = append(result, moduleMiddles...)
	result = append(result, inlineMiddles...)
	return result
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {
	msg := fmt.Sprintf("%s %s not found\n", req.Method, req.URL.Path)
	log.Errorf(msg)
	http.Error(w, msg, http.StatusNotFound)
}

func logRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("%s %s\n", req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}
