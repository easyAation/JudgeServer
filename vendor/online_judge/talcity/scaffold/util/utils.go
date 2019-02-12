package util

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/satori/go.uuid"
)

const (
	UUIDLength = 22
)

var (
	ShortFormat = "2006-01-02"
	LongFormat  = "2006-01-02 15:04:05.999"
)

// UUID return 22 chars uuid
func UUID() string {
	uid := uuid.NewV4()
	return base64.RawURLEncoding.EncodeToString(uid.Bytes())
}

// DeDuplicate 字符串数组去重
// Deprecated: Use values.DeDuplicateStrs instead.
func DeDuplicate(arr []string) []string {
	length := len(arr)
	if length < 2 {
		return arr
	}
	dic := make(map[string]struct{}, length)
	for _, s := range arr {
		dic[s] = struct{}{}
	}
	results := make([]string, 0, len(dic))
	for s := range dic {
		results = append(results, s)
	}
	return results
}

// DeDuplicateInt64 int64数组去重
// Deprecated: Use values.DeDuplicateInt64 instead.
func DeDuplicateInt64(arr []int64) []int64 {
	length := len(arr)
	if length < 2 {
		return arr
	}
	dic := make(map[int64]struct{}, length)
	for _, s := range arr {
		dic[s] = struct{}{}
	}
	results := make([]int64, 0, len(dic))
	for s := range dic {
		results = append(results, s)
	}
	return results
}

// DumpVal  dump interface to json format
func DumpVal(vals ...interface{}) {
	for _, val := range vals {
		prettyJSON, err := json.MarshalIndent(val, "", "    ")
		if err != nil {
			log.Println("dump err: ", err)
			return
		}
		log.Println(string(prettyJSON))
	}
}

// UserIP get user ip by read http request header
func UserIP(r *http.Request) string {
	var (
		// x-forward-for 的格式一般是 client_ip,proxy_ip,proxy_ip,..，需要截取，如果没有走代理，会是空的
		keys   = []string{"x-real-ip", "X-Real-Ip", "X-Forwarded-For", "x-forwarded-for", "remote_addr", "Remote_addr"}
		readIP = func(keys []string) string {
			for _, key := range keys {
				ip := r.Header.Get(key)
				if ip != "" {
					if strings.Contains(ip, ",") {
						ip = strings.Split(ip, ",")[0]
					}

					return ip
				}
			}

			return r.RemoteAddr
		}
	)

	// userIP = r.Header.Get("x-real-ip") || r.Header.Get("X-Real-Ip") || r.Header.Get("X-Forwarded-For") || r.Header.Get("x-forwarded-for") || r.Header.Get("remote_addr") || r.RemoteAddr
	userIP := readIP(keys)

	return userIP
}
