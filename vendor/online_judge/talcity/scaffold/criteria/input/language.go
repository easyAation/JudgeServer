package input

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const (
	httpHeaderLanguageKey = "Accept-Language"
	DefaultLanguage       = "EN-US"
	LanguageKey           = "language"
)

type LangQPair struct {
	Lang string
	Q    float64
}

func ParseAcceptLanguage(acceptLang string) []LangQPair {
	var results []LangQPair

	items := strings.Split(acceptLang, ",")
	for _, langQ := range items {
		langQ = strings.Trim(langQ, " ")
		if langQ == "" {
			continue
		}
		langPair := strings.Split(langQ, ";")
		if len(langPair) == 1 {
			results = append(results, LangQPair{langPair[0], 1})
		} else if len(langPair) == 2 {
			var (
				qValue float64
				err    error
			)
			qPair := strings.Split(langPair[1], "=")
			if len(qPair) >= 2 {
				if qValue, err = strconv.ParseFloat(qPair[1], 64); err != nil {
					qValue = 1
				}
			} else {
				qValue = 1
			}
			results = append(results, LangQPair{langPair[0], qValue})
		}
	}
	return results
}

// 从HTTP Header中取`Accept-Language`，并将其根据Q值进行稳定排序(降序)，返回结果中位于数组前面的语言是客户端更期望的
func LanguageFromHttpHeader(header http.Header) []string {
	value := header.Get(httpHeaderLanguageKey)
	lqs := ParseAcceptLanguage(value)
	sort.SliceStable(lqs, func(i, j int) bool {
		return lqs[i].Q > lqs[j].Q
	})
	languages := make([]string, 0, len(lqs))
	for _, item := range lqs {
		languages = append(languages, item.Lang)
	}
	return languages
}

func LanguageFromCtx(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	lgs, _ := ctx.Value(LanguageKey).([]string)
	return lgs
}

func LanguageWithDefault(ctx context.Context, default_ []string) []string {
	lgs := LanguageFromCtx(ctx)
	if len(lgs) == 0 {
		return default_
	}
	return lgs
}
