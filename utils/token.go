package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/easyAation/scaffold/db"

	"online_judge/JudgeServer/common"
)

func CreateToken(ctx context.Context, uid, password, auth string) (*string, error) {
	h := md5.New()
	if _, err := io.WriteString(h, uid); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(h, password); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(h, fmt.Sprintf("%d", time.Now().Unix())); err != nil {
		return nil, err
	}

	token := hex.EncodeToString(h.Sum(nil))
	if err := db.Set(ctx, uid, token, common.Config.Token.Expiration.D()); err != nil {
		return nil, err
	}
	return &token, nil
}

func GetToken(ctx context.Context, uid string) (string, error) {
	return db.GetStr(ctx, uid)
}
