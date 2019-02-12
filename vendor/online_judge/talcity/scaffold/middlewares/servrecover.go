package middlewares

import (
	"net/http"
	"runtime/debug"

	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/reply"
)

// BuildRecover wrap for recovering the panic.
func BuildRecover(moduleInternalErrCode int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					log.Errorf("server panic: %#v", v)
					debug.PrintStack()
					reply.WrapErr(nil, moduleInternalErrCode, "recover panic: %v", v)(w, r)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(f)
	}
}
