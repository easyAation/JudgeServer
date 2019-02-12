package middlewares

type AccessLogKey string

const (
	HTTPAccessLogKey AccessLogKey = "http_access_log"
	GrpcAccessLogKey AccessLogKey = "grpc_access_log"
)
