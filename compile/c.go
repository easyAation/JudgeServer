package compile

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const args = "gcc  -DONLINE_JUDGE  -O2  -w -fmax-errors=3  -std=c11 %s -lm -o %s"

type CCompile struct {
}

func (c *CCompile) Compile(codeDir, exeDir, codeName string) (string, error) {
	codeName = codeName + ".c"
	var (
		stderr bytes.Buffer
		codeFile = codeDir + string(filepath.Separator) + codeName
		exeFile  = exeDir + string(filepath.Separator) + codeName
	)
	cmd := exec.Command("bash", "-c", fmt.Sprintf(args, codeFile, exeFile))
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.WithMessage(err, stderr.String())
	}
	return exeFile, nil
}
