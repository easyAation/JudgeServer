package utils

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"

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
