package compile

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

const args = "gcc  -DONLINE_JUDGE  -O2  -w -fmax-errors=3  -std=c11 %s -lm -o %s"

type CCompile struct {
}

func (c *CCompile) Compile(srcPath, exePath string) (string, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", fmt.Sprintf(args, srcPath, exePath))
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.WithMessage(err, stderr.String())
	}
	return exePath, nil
}
