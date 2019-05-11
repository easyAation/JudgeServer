package utils

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
)

func Md5ForFile(fileName string) (string, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	fmt.Println("file context:--", string(data), "---")
	return fmt.Sprint(md5.Sum(data)), nil
}

func Md5ForString(message string) (string, error) {
	return fmt.Sprint(md5.Sum([]byte(message))), nil
}
func CovertMD5(nums [16]byte) string {
	var ans string
	for _, v := range nums {
		ans += strconv.Itoa(int(v))
	}
	return ans
}
