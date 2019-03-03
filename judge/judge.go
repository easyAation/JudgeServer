package judge

import "online_judge/JudgeServer/compile"

type Judge struct {
	compile   compile.Compile
	rpath     string
	exePath   string
	maxTime   int64
	maxMemory int64
}

type JudgeResponse struct {
}

type JudgeRequest struct {
	ID             string
	CodeContext    string
	Language       string
	maxTimeLimit   int64 // sec
	maxMemoryLimit int64
}

func (self *Judge) Judge() (JudgeResponse, error) {

}
