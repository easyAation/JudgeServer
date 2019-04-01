package route

import (
	"fmt"
	"online_judge/JudgeServer/common"
	"os"
	"path/filepath"

	"github.com/easyAation/scaffold/router"
)

func JudgeRouteModule() router.ModuleRoute {
	routes := []*router.Router{}

	return router.ModuleRoute{
		Routers: routes,
	}
}

// func judgeProblem(ctx *gin.Context) gin.HandlerFunc {
// 	var (
// 		request = struct {
// 			ID          string `json:"id"`
// 			ProblemID   int    `json:"problem_id"`
// 			Code        string `json:"code"`
// 			Language    string `json:"language"`
// 			TimeLimit   int64  `json:"time_limit"`
// 			MemoryLimit int64  `json:"memory_limit"`
// 		}{}
// 		codeFilePath string
// 	)
// 	err := ctx.ShouldBindJSON(&request)
// 	if err != nil {
// 		return reply.ErrorWithMessage(err, fmt.Sprintf("invalid param"))
// 	}
// 	if filepath.IsAbs(common.Config.Compile.CodeDir) {
// 		fmt.Println(common.Config.Compile.CodeDir, " is a abs path")
// 	} else {
// 		fmt.Println(common.Config.Compile.CodeDir, " not a abs path")
// 	}
// 	if Exists(common.Config.Compile.CodeDir) {
// 		fmt.Println(common.Config.Compile.CodeDir, " exist")
// 	} else {
// 		fmt.Println(common.Config.Compile.CodeDir, " not exist")
// 	}
// 	codeFilePath, err = createRequestFile(request.ID, request.Language, request.ProblemID)
// 	if err != nil {
// 		return reply.Err(err)
// 	}
// 	if err = ioutil.WriteFile(codeFilePath, []byte(request.Code), os.ModePerm); err != nil {
// 		return reply.Err(errors.WithStack(err))
// 	}
// 	fmt.Println("filaPath: ", codeFilePath)
// 	judge, err := judge.NewJudge(judge.Request{
// 		ID:          request.ID,
// 		ProblemID:   request.ProblemID,
// 		Code:        request.Code,
// 		Language:    request.Language,
// 		TimeLimit:   request.TimeLimit,
// 		MemoryLimit: request.MemoryLimit,
// 		FilePath:    codeFilePath,
// 	})
// 	if err != nil {
// 		return reply.Err(err)
// 	}
// 	response, err := judge.Run()
//
// 	if err != nil {
// 		return reply.Err(err)
// 	}
// 	return reply.Success(http.StatusOK, map[string]interface{}{
// 		"response": response,
// 	})
// }

func createRequestFile(rid, language string, problemid int) (string, error) {
	name := filepath.Join(common.Config.Compile.CodeDir, fmt.Sprintf("%s-%d", rid, problemid))
	switch language {
	case common.CLanguage:
		name += ".c"
	case common.CPPLanguage:
		name += ".cpp"
	case common.GoLanguage:
		name += ".go"
	}
	return name, nil
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
