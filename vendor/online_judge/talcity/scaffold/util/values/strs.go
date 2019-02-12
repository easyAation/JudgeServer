package values

// StrsIn tests whether string value in list
func StrsIn(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// StrsWithout remove values those presents in given list
func StrsWithout(list []string, values ...string) []string {
	if len(list) == 0 || len(values) == 0 {
		return list
	}
	var (
		r []string
		m = Strs2Map(values)
	)
	for _, v := range list {
		if _, ok := m[v]; !ok {
			r = append(r, v)
		}
	}
	return r
}

// Strs2Map converts list of strings to map, duplicated values will be removed
func Strs2Map(list []string) map[string]struct{} {
	var m = make(map[string]struct{})
	for _, v := range list {
		m[v] = struct{}{}
	}
	return m
}

// DeDuplicateStrs 字符串数组去重
func DeDuplicateStrs(arr []string) []string {
	length := len(arr)
	if length < 2 {
		return arr
	}
	dic := make(map[string]struct{}, length)
	for _, s := range arr {
		dic[s] = struct{}{}
	}
	result := make([]string, 0, len(dic))
	for s := range dic {
		result = append(result, s)
	}
	return result
}
