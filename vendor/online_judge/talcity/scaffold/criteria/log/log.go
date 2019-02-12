package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"

	"online_judge/talcity/scaffold/criteria/merr"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func init() {
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		DisableStacktrace: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error
	logger, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	sugar = logger.Sugar()
}

func Sync() {
	logger.Sync()
	sugar.Sync()
}

func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// Infow structed log method, arg keysAndValues MUST be key-value pair
func Infow(msg string, keysAndValues ...interface{}) {
	sugar.Infow(msg, keysAndValues...)
}

func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

// Errorf print msg and call stack
func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

// Panicf print msg and call stack and panic
func Panicf(template string, args ...interface{}) {
	sugar.Panicf(template, args...)
}

// Debug inline format, should only used in dev mode
func Debug(a ...interface{}) {
	_, fileName, line, _ := runtime.Caller(1)
	index := strings.LastIndex(fileName, "src/")
	if index > 0 {
		fileName = fileName[index+len("src/"):]
	}
	msg := fmtErrMsg(a...)
	fmt.Printf("%s:%d %s\n", fileName, line, msg)
}

// fmtErrMsg used to format error message
func fmtErrMsg(msgs ...interface{}) string {
	if len(msgs) > 1 {
		return fmt.Sprintf(msgs[0].(string), msgs[1:]...)
	}
	if len(msgs) == 1 {
		if v, ok := msgs[0].(string); ok {
			return v
		}
		if v, ok := msgs[0].(error); ok {
			return v.Error()
		}
	}
	return ""
}

// MarshalJSONOrDie should only used in dev mode
func MarshalJSONOrDie(v interface{}) string {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	return string(b)
}

// LogStructErr 结构化输出错误信息, 打印错误码和调用堆栈
func LogStructErr(msg string, keyAndValues map[string]interface{}, err error) {
	if err == nil {
		return
	}
	e := merr.Wrap(err, 0)
	keyAndValues["err code"] = e.Code
	keyAndValues["err msg"] = e.Msg
	keyAndValues["raw err"] = e.RawErr()
	keyAndValues["call stack"] = e.CallStack()
	args := make([]interface{}, 0, len(keyAndValues)*2)
	for k, v := range keyAndValues {
		args = append(args, k)
		args = append(args, v)
	}
	Infow(msg, args...)
}
