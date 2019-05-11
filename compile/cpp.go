package compile

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

type CPPCompile struct {
}

func (CPPCompile) Compile(codeFile, exePath string) (string, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("g++ -DONLINE_JUDGE -O2 -w -fmax-errors=3 -std=c++11 %s -lm -o %s", codeFile, exePath))
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.WithMessage(err, stderr.String())
	}
	return exePath, nil
}
