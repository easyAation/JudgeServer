package route

import (
	"fmt"
	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"net/http"
	"online_judge/JudgeServer/common"
	"path/filepath"
)

func ResourceRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter("/v1/image/upload",
			http.MethodPost,
			reply.Wrap(imageUpload),
		),
		router.NewRouter("/v1/image",
			http.MethodGet,
			reply.Wrap(getImage),
		),
	}
	return router.ModuleRoute{
		Routers: routes,
	}
}

func imageUpload(ctx *gin.Context) gin.HandlerFunc {
	imageFile, err := ctx.FormFile("image")
	if err != nil {
		return reply.Err(err)
	}
	url := filepath.Join(common.Config.Static.ImagePath, imageFile.Filename)
	fmt.Println(common.Config.Static.ImagePath)
	fmt.Println(url)
	err = ctx.SaveUploadedFile(imageFile, url)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"url": imageFile.Filename,
	})
}

func getImage(ctx *gin.Context) gin.HandlerFunc {
	imageName := ctx.Query("file")
	fmt.Println("image: ", imageName)
	return func(context *gin.Context) {
		ctx.File(filepath.Join(common.Config.Static.ImagePath, imageName))
	}
}
