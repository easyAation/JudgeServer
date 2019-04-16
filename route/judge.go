package route

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/easyAation/scaffold/db"
	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/model"
	"online_judge/JudgeServer/sandbox"
	"online_judge/JudgeServer/utils"
	"os"
	"path"
	"strconv"
	"strings"
)

func JudgeRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter(
			"/v1/judge_problem",
			http.MethodPost,
			reply.Wrap(judgeProblem),
		),
		router.NewRouter(
			"/v1/problem/add_data",
			http.MethodPost,
			reply.Wrap(addProblemData),
		),
	}

	return router.ModuleRoute{
		Routers: routes,
	}
}

func judgeProblem(ctx *gin.Context) gin.HandlerFunc {
	var request sandbox.Request
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		return reply.ErrorWithMessage(errors.WithStack(err), "invalid param")
	}
	sandBox, err := sandbox.NewSandBox(request)
	if err != nil {
		return reply.Err(err)
	}
	response, err := sandBox.Run()
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

func addProblemData(ctx *gin.Context) gin.HandlerFunc {
	pid := ctx.Query("pid")

	fmt.Println("pid: ", pid)
	dst := common.Config.SandBox.ProblemDir + string(os.PathSeparator) + pid
	os.MkdirAll(dst, os.ModePerm)
	fmt.Println("dst: ", dst)
	form, err := ctx.MultipartForm()
	if err != nil {
		return reply.Err(err)
	}
	files := form.File["files"]
	fileNameMp := make(map[string]int)
	var proDatas []model.ProblemData
	for _, file := range files {
		fileDir := path.Join(dst, FileNameNotExt(file.Filename))
		os.MkdirAll(fileDir, os.ModePerm)
		if err = ctx.SaveUploadedFile(file, path.Join(fileDir, file.Filename)); err != nil {
			return reply.Err(err)
		}
		fmt.Println(file.Filename)
		fileNameMp[strings.Split(file.Filename, ".")[0]] += 1
		if fileNameMp[strings.Split(file.Filename, ".")[0]] == 1 {
			pidInt, err := strconv.Atoi(pid)
			if err != nil {
				return reply.ErrorWithMessage(err, fmt.Sprintf("invali pid."))
			}
			proDatas = append(proDatas, model.ProblemData{
				PID:        pidInt,
				InputFile:  path.Join(fileDir, FileNameNotExt(file.Filename)+".in"),
				OutputFile: path.Join(fileDir, FileNameNotExt(file.Filename)+".out"),
			})
		}
	}

	for _, num := range fileNameMp {
		if num != 2 {
			return reply.Err(errors.Errorf("upload file format error."))
		}
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	for i := 0; i < len(proDatas); i++ {
		data, err := ioutil.ReadFile(proDatas[i].OutputFile)
		if err != nil {
			fmt.Println("hehehe", err)
			continue
		}
		proDatas[i].MD5 = utils.CovertMD5(md5.Sum(data))
		proDatas[i].MD5TrimSpace = utils.CovertMD5(md5.Sum(bytes.TrimSpace(data)))
	}
	lastId, err := model.AddProblemDatas(sqlExec, proDatas)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(http.StatusOK, map[string]interface{}{
		"data": lastId,
	})
}

func FileNameNotExt(name string) string {
	for i, c := range name {
		if c == '.' {
			return name[:i]
		}
	}
	return name
}

func convert(b []byte) string {
	s := make([]string, len(b))
	for i := range b {
		s[i] = strconv.Itoa(int(b[i]))
	}
	return strings.Join(s, ",")
}
