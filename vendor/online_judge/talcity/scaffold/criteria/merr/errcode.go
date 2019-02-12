package merr

import "strings"

/*
错误码
1. 错误码的作用是在微服务的环境中根据错误码判断出错误发生在哪个服务和错误类型
2. 公共组件会被每个服务引用, 无法通过公共组件中的错误码判断错误码发生于哪个服务
3. 为避免其他开发人员错误引用错误码,除"OK" 和系统错误码之外 merr不应该再声明任何错误码
   错误发生的位置仍然可以使用merr.Wrap 方法去包, 直接让错误码为0即可, 各个服务在调用公共组件时产生的错误都应该使用本服务特有错误码封装一下
4. 系统错误码 上游服务报错导致下游服务逻辑上不应该继续服务,此种情况应该用系统错误码表示, 系统错误码需要透传
5. 系统错误码占3位, 特定服务错误码占6位 其中前三位表示所说服务
   系统错误码由merr维护, 特定服务错误码由服务维护 并注册到errorMap
*/

const (
	// 系统错误码: 1xx
	AccountTokenInvalid = 100
	// 其他
	OK = 200
)

var errorMap = map[int]map[string]string{
	100: {
		"EN-US": "Account Token Not Valid",
		"ZH-CN": "登录状态已过期，请重新登录",
	},
	200: {
		"EN-US": "OK",
		"ZH-CN": "成功",
	},
}

// register error
func RegErr(errMap map[int]map[string]string) {
	for code, msgs := range errMap {
		errorMap[code] = msgs
	}
}

// get errMsg
func GetMsg(code int, languages []string) (string, bool) {
	msgMap, ok := errorMap[code]
	if !ok {
		return "Unknown Error", false
	}
	for _, lang := range languages {
		if msg, ok := msgMap[strings.ToUpper(lang)]; ok {
			if msg != "" {
				return msg, true
			}
		}
	}

	return "Unknown Error", false
}

// get all errorMap
func ErrMapping() map[int]map[string]string {
	return errorMap
}
