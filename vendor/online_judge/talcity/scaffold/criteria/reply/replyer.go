package reply

import (
	"encoding/json"
	"net/http"

	"online_judge/talcity/scaffold/criteria/input"
	"online_judge/talcity/scaffold/criteria/log"
	"online_judge/talcity/scaffold/criteria/merr"
)

/*
TODO:
中英文错误
*/

var showErrDescription bool = true

func SwitchRespErrDetail(swc bool) {
	showErrDescription = swc
}

type Response struct {
	Code        int         `json:"code"`
	Msg         string      `json:"msg,omitempty"`
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

func Wrap(f func(w http.ResponseWriter, r *http.Request) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responser := f(w, r)
		responser(w, r)
	}
}

func Success(content interface{}) http.HandlerFunc {
	return JSON(http.StatusOK, Response{
		Code: merr.OK,
		Data: content,
	})
}

func WrapErr(err error, code int, fmtAndArgs ...interface{}) http.HandlerFunc {
	return Err(merr.WrapDepth(1, err, code, fmtAndArgs...))
}

func Err(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e, ok := err.(*merr.MErr)
		if !ok {
			e = merr.Wrap(err, 0)
		}
		log.Debug(merr.ErrDetail(e))
		log.Errorf(merr.ErrDetail(e))

		var resp Response
		// TODO: language, 产品要求统一使用中文
		// 如果没有对应的错误码， 则使用e.Msg
		msg, _ := merr.GetMsg(e.Code, []string{"zh-cn"})
		resp = Response{
			Code: e.Code,
			Msg:  msg,
		}
		resp.Description = e.Msg

		responser := JSON(http.StatusBadRequest, resp)
		responser(w, r)
	}
}

func JSON(statusCode int, resp Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(statusCode)
		if resp.Msg == "" {
			// 使用middleware分析accept-language 存入context
			languages := input.LanguageWithDefault(r.Context(), []string{"zh-cn"})
			resp.Msg, _ = merr.GetMsg(resp.Code, languages)
		}
		if !showErrDescription {
			resp.Description = ""
		}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp); err != nil {
			// log err
			log.Errorf("encode response failed: %v", err)
			encoder.Encode(Response{
				Msg: err.Error(),
			})
		}
	}
}
