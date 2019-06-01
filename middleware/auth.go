package middleware

import (
	"log"
	
	"github.com/easyAation/scaffold/reply"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/utils"
)

const currentUser = "current_user"

func VerifyLogin(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader(common.AuthHeader)
		log.Println("token: ", token)
		if token == "" {
			reply.Err(errors.Errorf("invalid token. please login again."))(ctx)
			return
		}
		userID, err := utils.GetUIDByToken(ctx, token)
		if err != nil {
			reply.Err(err)(ctx)
			return
		}
		if token == "" || userID == "" {
			reply.Err(errors.Errorf("invalid token. please login again."))(ctx)
			return
		}
		ctx.Set(currentUser, userID)
		fn(ctx)
	}
}

func GetCurrentID(ctx *gin.Context) string {
	user, _ := ctx.Get(currentUser)
	switch user.(type) {
	case string:
		return user.(string)
	case []byte:
		return string(user.([]byte))
	default:
		return ""
	}
}
