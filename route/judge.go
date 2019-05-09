package route

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/easyAation/scaffold/db"
	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/model"
	"online_judge/JudgeServer/sandbox"
	"online_judge/JudgeServer/utils"
)

func JudgeRouteModule() router.ModuleRoute {
	routes := []*router.Router{
		router.NewRouter(
			"/v1/judge_problem",
			http.MethodPost,
			reply.Wrap(judgeProblem),
		),
		router.NewRouter(
			"/v1/submission/submit",
			http.MethodPost,
			reply.Wrap(judgeProblem),
		),
		router.NewRouter(
			"/v1/problem/add_data",
			http.MethodPost,
			reply.Wrap(addProblemData),
		),
		router.NewRouter(
			"/v1/problem/add",
			http.MethodPost,
			reply.Wrap(addProblem),
		),
		router.NewRouter(
			"/v1/problem/update",
			http.MethodPost,
			reply.Wrap(updateProblem),
		),
		router.NewRouter(
			"/v1/problem/detail",
			http.MethodGet,
			reply.Wrap(getProblem),
		),
		router.NewRouter(
			"/v1/problem/list",
			http.MethodGet,
			reply.Wrap(getProblems),
		),
		router.NewRouter(
			"/v1/submit/list",
			http.MethodGet,
			reply.Wrap(getSubmits),
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

	fmt.Println("request: ", request)
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	rowsAffected, err := model.AddSubmit(sqlExec, &model.Submit{
		PID:      request.ProblemID,
		SubmitID: request.ID,
		Code:     request.Code,
	})
	if err != nil {
		return reply.Err(err)
	}
	log.Printf("%d rows affected.", rowsAffected)
	sandBox, err := sandbox.NewSandBox(request)
	if err != nil {
		return reply.Err(err)
	}
	response, err := sandBox.Run()
	if err != nil {
		return reply.Err(err)
	}
	go func() {
		status := common.Accept
		sort.Slice(response, func(i, j int) bool {
			return response[i].Index < response[j].Index
		})
		for _, result := range response {
			if result.Status != common.Accept {
				status = result.Status
				break
			}
		}
		_, err := model.UpdateSubmitBySID(sqlExec, request.ID, map[string]interface{}{
			"result": status,
		})
		if err != nil {
			log.Println(err)
		}
	}()
	return reply.Success(http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

func addProblem(ctx *gin.Context) gin.HandlerFunc {
	var problem model.Problem
	err := ctx.ShouldBindJSON(&problem)
	if err != nil {
		return reply.ErrorWithMessage(err, "invalid param")
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	_, err = model.AddProblem(sqlExec, problem)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"id": problem.ID,
	})
}

func updateProblem(ctx *gin.Context) gin.HandlerFunc {
	var (
		problem = struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			TimeLimit   int64  `json:"time_limit"`
			MemoryLimit int64  `json:"memory"`
			Description string `json:"description"`
			InputDes    string `json:"input_des"`
			OutputDes   string `json:"output_des"`
			Input       string `json:"case_data_input"`
			Output      string `json:"case_data_output"`
		}{}
	)
	err := ctx.ShouldBindJSON(&problem)
	if err != nil {
		return reply.ErrorWithMessage(err, "invalid param")
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	_, err = model.UpdateProblem(sqlExec, problem.ID, map[string]interface{}{
		"name":             problem.Name,
		"time_limit":       problem.TimeLimit,
		"memory_limit":     problem.MemoryLimit,
		"description":      problem.Description,
		"input_des":        problem.InputDes,
		"output_des":       problem.OutputDes,
		"case_data_input":  problem.Input,
		"case_data_output": problem.Output,
	})
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(http.StatusOK, nil)
}
func getProblem(ctx *gin.Context) gin.HandlerFunc {
	ctx.Header("Access-Control-Allow-Origin", "*")
	pid := ctx.Query("pid")
	if pid == "" {
		return reply.Err(errors.Errorf("invalid param pid: %v", pid))
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), model.ProblemTable)
	if err != nil {
		return reply.Err(err)
	}
	problem, err := model.GetOneProblem(sqlExec, map[string]interface{}{
		"id": pid,
	})
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"problem": problem,
	})
}

func getProblems(ctx *gin.Context) gin.HandlerFunc {
	ctx.Header("Access-Control-Allow-Origin", "*")
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), model.ProblemTable)
	if err != nil {
		return reply.Err(err)
	}
	problemList, err := model.GetProblem(sqlExec, nil)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"list": problemList,
	})
}

func addProblemData(ctx *gin.Context) gin.HandlerFunc {
	pid := ctx.Query("pid")
	dst := common.Config.SandBox.ProblemDir + string(os.PathSeparator) + pid
	os.MkdirAll(dst, os.ModePerm)
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

// getSubmits support filters of uid, pid, language
func getSubmits(ctx *gin.Context) gin.HandlerFunc {
	ctx.Header("Access-Control-Allow-Origin", "*")
	pid := ctx.Query("pid")
	language := ctx.Query("language")

	var filters map[string]interface{}
	if pid != "" || language != "" {
		filters = make(map[string]interface{})
	}
	if pid != "" {
		filters["pid"] = pid
	}
	if language != "" {
		filters["language"] = language
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	submits, err := model.GetSubmits(sqlExec, filters)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"data": submits,
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
