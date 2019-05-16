package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/easyAation/scaffold/db"

	"online_judge/JudgeServer/common"
)

func CreateToken(ctx context.Context, uid, password, auth string) (*string, error) {
	token := EncryptPassword(uid, password, fmt.Sprintf("%d", time.Now().Unix()))
	if err := db.Set(ctx, uid, token, common.Config.Token.Expiration.D()); err != nil {
		return nil, err
	}
	if err := db.Set(ctx, token, uid, common.Config.Token.Expiration.D()); err != nil {
		return nil, err
	}
	return &token, nil
}

func GetTokenByUID(ctx context.Context, uid string) (string, error) {
	return db.GetStr(ctx, uid)
}

func GetUIDByToken(ctx context.Context, token string) (string, error) {
	return db.GetStr(ctx, token)
}
