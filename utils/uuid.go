package utils

import (
	"encoding/base64"
	uuid2 "github.com/satori/go.uuid"
)

func UUID() string {
	uid, _ := uuid2.NewV4()
	return base64.RawURLEncoding.EncodeToString(uid.Bytes())
}
