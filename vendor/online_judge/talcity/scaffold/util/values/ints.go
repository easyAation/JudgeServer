package values

// IntsIn tests whether int value in list
func IntsIn(list []int, value int) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// Int64sIn tests whether int64 value in list
func Int64sIn(list []int64, value int64) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// DeDuplicateInt64 int64数组去重
func DeDuplicateInt64(arr []int64) []int64 {
	length := len(arr)
	if length < 2 {
		return arr
	}
	dic := make(map[int64]struct{}, length)
	for _, s := range arr {
		dic[s] = struct{}{}
	}
	result := make([]int64, 0, len(dic))
	for s := range dic {
		result = append(result, s)
	}
	return result
}
