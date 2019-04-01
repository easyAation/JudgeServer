package judge

import (
	"fmt"

	"online_judge/JudgeServer/common"
	"online_judge/JudgeServer/compile"
)

type Judge struct {
	compile.Compiler
	Request
}

type Response struct {
}

type Request struct {
	ID          string `json:"id"`
	ProblemID   int    `json:"problem_id"`
	Code        string `json:"code"`
	FilePath    string `json:"-"`
	Language    string `json:"language"`
	TimeLimit   int64  `json:"time_limit"` // nsec
	MemoryLimit int64  `json:"memory_limit"`
}

func NewJudge(request Request) (*Judge, error) {
	compile, err := compile.NewCompile(request.Language)
	if err != nil {
		return nil, err
	}
	judge := Judge{
		Compiler: compile,
		Request:  request,
	}
	return &judge, nil
}

func (judge *Judge) Run() (*Response, error) {
	execFile, err := judge.Compile(common.Config.Compile.CodeDir, common.Config.Compile.ExeDir, fmt.Sprintf("%s-%d", judge.ID, judge.ProblemID))
	if err != nil {
		return nil, err
	}
	fmt.Printf(execFile)
	return nil, nil
}
