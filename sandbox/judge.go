package sandbox

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/compile"
	"online_judge/JudgeServer/model"
	"online_judge/JudgeServer/utils"
)

type SandBox struct {
	compile.Compiler
	Request
	codeFile string
	exeFile  string
}
type Result struct {
	Index  int
	Time   int64 `json:"real_time"`
	Memory int64 `json:"memory"`
	Code   int   `json:"result"`
	Status string
}

type Request struct {
	ID          string `json:"id"`
	ProblemID   int    `json:"problem_id"`
	Code        string `json:"code"`
	Language    string `json:"language"`
	TimeLimit   int64  `json:"time_limit"` // nsec
	MemoryLimit int64  `json:"memory_limit"`
}

func judge(code int, file1 string, proData model.ProblemData) string {
	if code == 1 || code == 2 {
		return common.TimeLimit
	}
	if code == 3 {
		return common.MemoryLimit
	}
	if code == 4 {
		return common.MemoryLimit
	}
	if code == 5 {
		return common.SysteamError
	}
	if code != 0 {
		return common.InternalError
	}

	data, err := ioutil.ReadFile(file1)
	if err != nil {
		return common.InternalError
	}

	if utils.CovertMD5(md5.Sum(data)) != proData.MD5 {
		if utils.CovertMD5(md5.Sum(bytes.TrimSpace(data))) == proData.MD5TrimSpace {
			return common.PresentationError
		}
		return common.WrongAnswer
	}
	return common.Accept

}
func buildCommandArgs(values map[string]interface{}) string {
	var args = make([]string, 0, len(values))
	for op, value := range values {
		args = append(args, fmt.Sprintf("--%s=%v", op, value))
	}
	return " " + strings.Join(args, " ")
}

func NewSandBox(request Request) (*SandBox, error) {
	compile, err := compile.NewCompile(request.Language)
	if err != nil {
		return nil, err
	}
	return &SandBox{
		Compiler: compile,
		Request:  request,
	}, nil
}

func (s *SandBox) SaveCodeFile() error {
	s.codeFile = filepath.Join(common.Config.Compile.CodeDir, fmt.Sprintf("%s_%d", s.ID, s.ProblemID))
	switch s.Language {
	case common.CLanguage:
		s.codeFile += ".c"
	case common.CPPLanguage:
		s.codeFile += ".cpp"
	case common.GoLanguage:
		s.codeFile += ".go"
	default:
		return errors.Errorf("%s not support.", s.Language)
	}
	if err := ioutil.WriteFile(s.codeFile, []byte(s.Code), os.ModePerm); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SandBox) compile() error {
	if s.exeFile != "" {
		return nil
	}
	var err error
	s.exeFile, err = s.Compile(s.codeFile, common.Config.Compile.ExeDir+string(os.PathSeparator)+s.ID)
	if err != nil {
		return err
	}
	return nil
}
func (s *SandBox) Run() ([]Result, error) {
	if err := s.SaveCodeFile(); err != nil {
		return nil, errors.Wrap(err, "save file error.")
	}
	if err := s.compile(); err != nil {
		return nil, errors.Wrap(err, "compile error.")
	}

	sqlExec, err := db.GetSqlExec(context.Background(), "problem")
	if err != nil {
		return nil, errors.Wrap(err, "get sqlExec error.")
	}

	problemData, err := model.GetProblemData(sqlExec, map[string]interface{}{
		"pid": s.ProblemID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	results := make([]Result, 0, len(problemData))
	for index, prodata := range problemData {
		outputFile := common.Config.SandBox.OutPutDir + string(os.PathSeparator) + s.ID + fmt.Sprintf("_%d", index)
		args := common.Config.SandBox.Exe + buildCommandArgs(map[string]interface{}{
			"exe_path":          s.exeFile,
			"input_path":        prodata.InputFile,
			"output_path":       outputFile,    // outputFile,
			"max_cpu_time":      s.TimeLimit,   // s.TimeLimit,
			"max_real_time":     s.TimeLimit,   // s.TimeLimit,
			"memory_limit":      s.MemoryLimit, // s.MemoryLimit,
			"seccomp_rule_name": "c_cpp",
		})
		cmd := exec.Command("/usr/bin/bash", "-c", args)
		msg, err := cmd.CombinedOutput()
		fmt.Println(string(msg))
		if err != nil {
			return nil, errors.Wrap(err, string(msg))
		}
		var result = Result{
			Index: index,
		}
		if err := json.Unmarshal(msg, &result); err != nil {
			log.Print(err)
			continue
		}
		result.Status = judge(result.Code, outputFile, prodata)
		results = append(results, result)
		fmt.Printf("output file: %s\n", outputFile)
	}
	return results, nil
}
