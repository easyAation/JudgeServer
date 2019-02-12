package middlewares

import (
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/merr"
)

type GrpcAccessLog struct {
	Method   string `protobuf:"bytes,1,opt,name=method" json:"method,omitempty"`
	ErrMsg   string `protobuf:"bytes,2,opt,name=err_msg,json=errMsg" json:"err_msg,omitempty"`
	Code     int    `protobuf:"varint,3,opt,name=code" json:"code,omitempty"`
	Start    int64  `protobuf:"varint,4,opt,name=start" json:"start,omitempty"`
	Cost     string `protobuf:"bytes,5,opt,name=cost" json:"cost,omitempty"`
	ClientIP string `protobuf:"bytes,6,opt,name=client_ip,json=client_ip" json:"client_ip,omitempty"`
	Request  string `protobuf:"bytes,7,opt,name=request" json:"request,omitempty"`
	Response string `protobuf:"bytes,8,opt,name=response" json:"response,omitempty"`
}

func BuildUnaryLogger(panicErrCode, internalErrCode int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		method := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		var clientIP string
		if peer, ok := peer.FromContext(ctx); ok {
			clientIP = peer.Addr.String()
		}
		reqBody := ""
		if a, ok := respToString(req); ok {
			reqBody = a
		}

		accessLog := &GrpcAccessLog{
			Method:   method,
			Start:    time.Now().UnixNano(),
			ClientIP: clientIP,
			Request:  reqBody,
		}

		defer func() {
			cost := time.Now().Sub(time.Unix(0, accessLog.Start))
			accessLog.Cost = cost.String()

			if a, ok := respToString(resp); ok {
				accessLog.Response = a
			}

			if e := recover(); e != nil {
				me := merr.Wrap(nil, panicErrCode, "panic: %v", e)
				accessLog.Code = me.Code
				accessLog.ErrMsg = me.Msg
				err = me
			} else if err != nil {
				if s, ok := status.FromError(err); ok {
					accessLog.Code = int(s.Code())
					accessLog.ErrMsg = s.Message()
				} else if me, ok := err.(*merr.MErr); ok {
					accessLog.Code = me.Code
					accessLog.ErrMsg = me.Msg
				} else {
					accessLog.Code = internalErrCode
					accessLog.ErrMsg = err.Error()
				}
			}

			log.Infow("oss grpc access log",
				"method", accessLog.Method,
				"errMsg", accessLog.ErrMsg,
				"code", accessLog.Code,
				"start", accessLog.Start,
				"cost", accessLog.Cost,
				"client_ip", accessLog.ClientIP,
				"request", accessLog.Request,
				"resp", accessLog.Response,
			)
		}()
		ctx = context.WithValue(ctx, GrpcAccessLogKey, accessLog)
		resp, err = handler(ctx, req)
		return
	}
}

func respToString(resp interface{}) (string, bool) {
	data, err := json.Marshal(resp)
	if err == nil {
		return string(data), true
	}

	if a, ok := resp.(stringAble); ok {
		return a.String(), true
	}
	return "", false
}

type stringAble interface {
	String() string
}

// GetGrpcAccessLog 必须使用了 UnaryLogger 才能调用此函数, 否则判定为程序逻辑错误, 应该panic
func GetGrpcAccessLog(ctx context.Context) *GrpcAccessLog {
	return ctx.Value(GrpcAccessLogKey).(*GrpcAccessLog)
}
