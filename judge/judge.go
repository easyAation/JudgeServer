package judge

import (
	"online_judge/JudgeServer/compile"
)

type Judge struct {
	Compile     compile.Compile
	ResoucePath string
	ExePath     string
	Request     JudgeRequest
}

type JudgeResponse struct {
}

type JudgeRequest struct {
	ID             string
	problemID      int
	CodeContext    string
	Language       string
	maxTimeLimit   int64 // sec
	maxMemoryLimit int64
}

func NewJudge(request JudgeRequest) (*Judge, error) {
	compile, err := compile.NewCompile(request.Language)
	if err != nil {
		return nil, err
	}
	judge := Judge{
		Compile: compile,
		Request: request,
	}
	return &judge, nil
}

func (self *Judge) Judge() (JudgeResponse, error) {

}
