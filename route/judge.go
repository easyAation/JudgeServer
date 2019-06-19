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
	"time"

	"github.com/easyAation/scaffold/db"
	"github.com/easyAation/scaffold/reply"
	"github.com/easyAation/scaffold/router"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/middleware"
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
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/submission/submit",
			http.MethodPost,
			reply.Wrap(judgeProblem),
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/problem/add_data",
			http.MethodPost,
			reply.Wrap(addProblemData),
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/problem/add",
			http.MethodPost,
			reply.Wrap(addProblem),
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/problem/update",
			http.MethodPost,
			reply.Wrap(updateProblem),
			middleware.VerifyLogin,
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
		router.NewRouter(
			"/v1/user/solves",
			http.MethodGet,
			reply.Wrap(getUserSolve),
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/contest/create",
			http.MethodPost,
			reply.Wrap(createContest),
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/contest",
			http.MethodGet,
			reply.Wrap(getContest),
		),
		router.NewRouter(
			"/v1/contest/submit",
			http.MethodPost,
			reply.Wrap(judgeContestSubmit),
			middleware.VerifyLogin,
		),
		router.NewRouter(
			"/v1/contest/submission",
			http.MethodGet,
			reply.Wrap(contestStatus),
		),
		router.NewRouter("/v1/contest/rank",
			http.MethodGet,
			reply.Wrap(contestRank),
		),
		router.NewRouter("/v1/contest/list",
			http.MethodGet,
			reply.Wrap(contestList),
		),
	}

	return router.ModuleRoute{
		Routers: routes,
	}
}

func contestRank(ctx *gin.Context) gin.HandlerFunc {
	type (
		problem = struct {
			PID      int       `json:"pid"`
			Failed   int       `json:"failed"`
			CreateAT time.Time `json:"create_at"`
		}
		user = struct {
			ID       string    `json:"id"`
			Name     string    `json:"name"`
			Problems []problem `json:"problems"`
		}
	)
	cid := ctx.Query("cid")
	if cid == "" {
		return reply.Success(200, nil)
	}
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return reply.Err(err)
	}
	allSubmit, err := model.GetContestSubmit(sqlExec, map[string]interface{}{
		"cid": cid,
	})
	if err != nil {
		return reply.Err(err)
	}
	allAccount, err := model.GetAccounts(ctx, nil)
	if err != nil {
		return reply.Err(err)
	}
	acs := make(map[string]string)
	for _, ac := range allAccount {
		acs[ac.ID] = ac.Name
	}
	Rank := make(map[string]*user)
	for _, sb := range allSubmit {
		if _, ok := Rank[sb.UID]; !ok {
			Rank[sb.UID] = &user{
				sb.UID,
				acs[sb.UID],
				make([]problem, 0),
			}
		}
	}
	var found bool
	var pro *problem
	for _, sb := range allSubmit {
		pro = nil
		found = false
		if Rank[sb.UID].ID == sb.UID {
			found = true
		}
		if !found {
			continue
		}
		for i := 0; i < len(Rank[sb.UID].Problems); i++ {
			if Rank[sb.UID].Problems[i].PID == sb.PID {
				pro = &Rank[sb.UID].Problems[i]
			}
		}
		if pro == nil {
			failed := 0
			if sb.Result != common.Accept {
				failed = -1
			}
			Rank[sb.UID].Problems = append(Rank[sb.UID].Problems, problem{
				sb.PID,
				failed,
				sb.CreatedAT,
			})

		} else {
			if pro.CreateAT.After(sb.CreatedAT) {
				pro.CreateAT = sb.CreatedAT
			}
			if pro.Failed < 0 {
				if sb.Result != common.Accept {
					pro.Failed--
				} else {
					pro.Failed = -pro.Failed
				}
			}
		}
	}
	return func(context *gin.Context) {
		context.JSON(200, Rank)
	}
}

func judgeContestSubmit(ctx *gin.Context) gin.HandlerFunc {
	var (
		request = struct {
			sandbox.Request
			CID int64 `json:"cid"`
		}{}
	)
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		return reply.ErrorWithMessage(err, "invalid param")
	}
	fmt.Printf("%+v\n", request)
	sandBox, err := sandbox.NewSandBox(request.Request)
	if err != nil {
		return reply.Err(err)
	}
	res, err := sandBox.Run()
	if err != nil {
		return reply.Err(err)
	}

	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return reply.Err(err)
	}
	// log.Println(middleware.GetCurrentID(ctx))
	_, err = model.AddContestSubmit(sqlExec, model.ContestSubmit{
		CID: request.CID,
		Submit: model.Submit{
			PID:      request.ProblemID,
			UID:      middleware.GetCurrentID(ctx),
			SubmitID: request.ID,
			Code:     request.Code,
			Language: request.Language,
			Result:   res.Status,
			RunTime:  res.Time,
			Memory:   res.Memory,
		},
	})
	if err != nil {
		return reply.Err(err)
	}

	return reply.Success(200, map[string]interface{}{
		"data": struct {
			Result string `json:"result"`
			Time   int64  `json:"time"`
			Memory int64  `json:"memory"`
		}{
			res.Status,
			res.Time,
			res.Memory,
		},
	})
}

func judgeProblem(ctx *gin.Context) gin.HandlerFunc {
	var request sandbox.Request
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		return reply.ErrorWithMessage(errors.WithStack(err), "invalid param")
	}

	fmt.Println("request: ", request)

	sandBox, err := sandbox.NewSandBox(request)
	if err != nil {
		return reply.Err(err)
	}
	res, err := sandBox.Run()
	if err != nil {
		return reply.Err(err)
	}
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}

	rowsAffected, err := model.AddSubmit(sqlExec, &model.Submit{
		PID:      request.ProblemID,
		UID:      middleware.GetCurrentID(ctx),
		SubmitID: request.ID,
		Code:     request.Code,
		Language: request.Language,
		Result:   res.Status,
		RunTime:  res.Time,
		Memory:   res.Memory,
	})
	if err != nil {
		return reply.Err(err)
	}
	log.Printf("%d rows affected.", rowsAffected)

	return reply.Success(http.StatusOK, map[string]interface{}{
		"data": struct {
			Result string `json:"result"`
			Time   int64  `json:"tint"`
			Memory int64  `json:"memory"`
		}{
			res.Status,
			res.Time,
			res.Memory,
		},
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
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
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

func contestList(ctx *gin.Context) gin.HandlerFunc {
	sqlExec, err := db.GetSqlExec(ctx.Request.Context(), "problem")
	if err != nil {
		return reply.Err(err)
	}
	contestList, err := model.GetContest(sqlExec, nil)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"list": contestList,
	})
}

func contestStatus(ctx *gin.Context) gin.HandlerFunc {
	cid := ctx.Query("cid")
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return reply.Err(err)
	}
	var filter map[string]interface{}
	if cid != "" {
		filter = make(map[string]interface{})
		filter["cid"] = cid
	}
	submits, err := model.GetContestSubmit(sqlExec, filter)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"list":  submits,
		"total": len(submits),
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
	sid := ctx.Query("sid")
	pid := ctx.Query("pid")
	language := ctx.Query("language")

	var filters map[string]interface{}
	if pid != "" || language != "" || sid != "" {
		filters = make(map[string]interface{})
	}
	if sid != "" {
		filters["submit_id"] = sid
	}
	if pid != "" {
		filters["pid"] = pid
	}
	if language != "" {
		filters["language"] = language
	}
	submits, err := model.GetSubmits(ctx, filters)
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"data":  submits,
		"total": len(submits),
	})
}

func getUserSolve(ctx *gin.Context) gin.HandlerFunc {
	pids, err := model.GetUserSolves(ctx, middleware.GetCurrentID(ctx))
	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"data": pids,
	})
}

func createContest(ctx *gin.Context) gin.HandlerFunc {
	var (
		c = struct {
			Title      string `json:"title"`
			Encrypt    int    `json:"encrypt"`
			StartAt    int64  `json:"start"`
			EndAt      int64  `json:"end"`
			ProblemIDs []int  `json:"list"`
		}{}
	)
	err := ctx.ShouldBindJSON(&c)
	if err != nil {
		return reply.Err(errors.Wrap(err, ""))
	}
	fmt.Println(c)
	cid, err := model.AddContest(ctx, model.Contest{
		Title:   c.Title,
		Encrypt: c.Encrypt,
		StartAt: time.Unix(c.StartAt/1000, c.StartAt%1000),
		EndAt:   time.Unix(c.EndAt/1000, c.StartAt%1000),
	})
	if err != nil {
		return reply.Err(err)
	}

	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return reply.ErrorWithMessage(err, "internal error")
	}
	err = func() error {
		tx, err := sqlExec.Beginx()
		if err != nil {
			return errors.Errorf("internal error")
		}
		defer func() {
			if err != nil && tx != nil {
				tx.Rollback()
			}
		}()
		for index, pid := range c.ProblemIDs {
			_, err = tx.Exec("INSERT INTO contest_problem (cid,pid, position) VALUES (?, ?, ?)", cid,
				pid, index)
			if err != nil {
				return errors.Errorf("internal error")
			}
		}
		return tx.Commit()
	}()

	if err != nil {
		return reply.Err(err)
	}
	return reply.Success(200, map[string]interface{}{
		"cid": cid,
	})
}

func getContest(ctx *gin.Context) gin.HandlerFunc {
	cid := ctx.Query("cid")
	if cid == "" {
		return reply.ErrorWithMessage(nil, "invalid cid")
	}
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return reply.Err(err)
	}
	contest, err := model.GetOneContest(sqlExec, map[string]interface{}{
		"id": cid,
	})
	if err != nil {
		return reply.Err(err)
	}
	list, err := model.GetContestProblems(sqlExec, map[string]interface{}{
		"cid": cid,
	})
	if err != nil {
		return reply.Err(err)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Position < list[j].Position
	})

	allProblems, err := model.GetProblem(sqlExec, nil)
	if err != nil {
		return reply.Err(err)
	}

	for i := 0; i < len(list); i++ {
		for _, pro := range allProblems {
			if list[i].PID == pro.ID {
				list[i].Title = pro.Name
				break
			}
		}
	}
	return reply.Success(200, map[string]interface{}{
		"contest": contest,
		"list":    list,
		"total":   len(list),
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
