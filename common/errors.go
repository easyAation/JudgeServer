package common

import "online_judge/talcity/scaffold/criteria/merr"

const (
	InvalidParam = 101000
	InternalError = 101001
)

func init() {
	merr.RegErr(map[int]map[string]string{
		101000: map[string]string{
			"EN-US": "Invalid Param",
			"ZH-CN": "参数格式错误，无效的参数",
		},
		101001: map[string]string{
			"EN-US": "Internal Error",
			"ZH-CN": "内部错误",
		},
	})
}