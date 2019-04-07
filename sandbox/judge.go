package sandbox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/compile"
	"online_judge/JudgeServer/model"
	"online_judge/JudgeServer/utils"
)

const (
	accept        = "Accept"
	wrongAnswer   = "Wrong Answer"
	timeLimit     = "Time Limit"
	memoryLimit   = "Memory Limit"
	runtimeError  = "Runtime Error"
	systeamError  = "System Error"
	internalError = "internal Error"
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

func CodeToStatus(code int) string {
	if code == 1 || code == 2 {
		return timeLimit
	}
	if code == 3 {
		return memoryLimit
	}
	if code == 4 {
		return runtimeError
	}
	if code == 5 {
		return systeamError
	}
	return internalError
}

func judge(code int, file1, file2 string) string {
	if code != 0 {
		return CodeToStatus(code)
	}
	data1, err := utils.Md5ForFile(file1)
	if err != nil {
		return internalError
	}
	data2, err := utils.Md5ForFile(file2)
	if err != nil {
		return internalError
	}
	fmt.Println("data1: ", data1)
	fmt.Println("data2: ", data2)
	if data1 == data2 {
		return accept
	}
	return wrongAnswer
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

	problem, err := model.GetProBlemByID(s.ProblemID)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	results := make([]Result, 0, len(problem.IOFiles))
	for index, ioFile := range problem.IOFiles {
		outputFile := common.Config.SandBox.OutPutDir + string(os.PathSeparator) + s.ID + fmt.Sprintf("_%d", index)
		args := common.Config.SandBox.Exe + buildCommandArgs(map[string]interface{}{
			"exe_path":          s.exeFile,
			"input_path":        ioFile.InputFile,
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
		result.Status = judge(result.Code, ioFile.OutputFile, outputFile)
		results = append(results, result)
		fmt.Printf("output file: %s\n", outputFile)
	}
	return results, nil
}
