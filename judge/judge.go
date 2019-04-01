package judge

import (
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

func (r Judge) Run() {

}
