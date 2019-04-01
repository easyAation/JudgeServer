package compile

import (
	"fmt"

	"github.com/pkg/errors"

	"online_judge/JudgeServer/common"
)

type Compile interface {
	Compile(srcPath, execPath string) (string, error)
}

func NewCompile(language string) (Compile, error) {
	switch language {
	case common.CLanguage:
		return &CCompile{}, nil
	case common.CPPLanguage:
		return &CPPCompile{}, nil

	}
	return nil, errors.New(fmt.Sprintf("%s not found.", language))
}
