package caller

import (
	"context"
	"net/http"
)

const (
	callerName = "Caller-Name"
)

type callerNameKey struct{}
type ServerMiddleware func(rw http.ResponseWriter, req *http.Request, next http.Handler)

func ContextWithCallerName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, callerNameKey{}, name)
}

func CallerNameFromContext(ctx context.Context) string {
	name, ok := ctx.Value(callerNameKey{}).(string)
	if !ok {
		return ""
	}
	return name
}

// ExtractCallerName 从header中获取接口调用者名称用于监控信息收集
func ExtractCallerName() ServerMiddleware {
	return func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		name := req.Header.Get(callerName)
		next.ServeHTTP(rw, req.WithContext(ContextWithCallerName(req.Context(), name)))
	}
}
