package middlewares

import (
	"bytes"
	"context"
	"mime"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-chi/chi/middleware"

	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/reply"
	"online_judge/talcity/scaffold/criteria/trace"
)

type HTTPAccessLog struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Request    string `json:"request,omitempty"`
	StatusCode int    `json:"status_code"`
	Response   string `json:"response,omitempty"`
	ClientIP   string `json:"client_ip"`
	Cost       string `json:"cost"`
}

func BuildHttpLogger(internalErrCode int, msg string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			buf := new(bytes.Buffer)
			ww.Tee(buf)

			var doubtfulMultipartForm = false
			if r.Method == http.MethodPut || r.Method == http.MethodPost {
				v := r.Header.Get("Content-Type")
				if v == "" {
					doubtfulMultipartForm = false
				} else {
					d, _, _ := mime.ParseMediaType(v)
					doubtfulMultipartForm = d == "multipart/form-data"
				}
			}
			dumpRequest, err := httputil.DumpRequest(r, !doubtfulMultipartForm)
			if err != nil {
				reply.WrapErr(err, internalErrCode, "dump request failed")(w, r)
				return
			}

			start := time.Now()
			accessLog := &HTTPAccessLog{
				Method:   r.Method,
				Path:     r.RequestURI,
				Request:  string(dumpRequest),
				ClientIP: r.RemoteAddr,
			}
			ctx := context.WithValue(r.Context(), HTTPAccessLogKey, accessLog)
			// set trace
			tr := trace.New()
			ctx = trace.ToCtx(ctx, tr)

			defer func() {
				accessLog.StatusCode = int(ww.Status())
				accessLog.Response = string(buf.Bytes())
				accessLog.Cost = time.Now().Sub(start).String()

				keys, trs := tr.Traces()
				args := make([]interface{}, 0, (len(keys)+7)*2)
				args = append(args,
					"method", accessLog.Method,
					"path", accessLog.Path,
					"request", accessLog.Request,
					"status_code", accessLog.StatusCode,
					"response", accessLog.Response,
					"client_ip", accessLog.ClientIP,
					"cost", accessLog.Cost)
				for _, k := range keys {
					args = append(args, k, trs[k])
				}

				log.Infow(msg, args...)
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// GetHTTPAccessLog 必须使用了HTTPLogger 才能调用此函数, 否则判定为程序逻辑错误, 应该panic
func GetHTTPAccessLog(r *http.Request) *HTTPAccessLog {
	return r.Context().Value(HTTPAccessLogKey).(*HTTPAccessLog)
}
