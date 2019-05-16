package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
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

func EncryptPassword(args ...string) string {
	return MD5Sum(args...)
}

func MD5Sum(args ...string) string {
	h := md5.New()
	for _, arg := range args {
		io.WriteString(h, arg)
	}
	return hex.EncodeToString(h.Sum(nil))
}
