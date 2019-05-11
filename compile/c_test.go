package compile

import (
	"log"
	"os"
	"testing"

	"online_judge/JudgeServer/common"
)

const codeDir = "~/.online_judge/code_path"
const exeDir = "~/.online_judge/exe_path"

func TestCompile(t *testing.T) {
	mkdirIfNotExist(codeDir, t)
	mkdirIfNotExist(exeDir, t)

	c, err := NewCompile(common.CLanguage)
	if err != nil {
		t.Error(err)
	}
	exeFile, err := c.Compile(codeDir+"/hello.c", exeDir+"/hello")
	if err != nil {
		t.Error(err)
	}
	log.Print(exeFile)

	cpp, err := NewCompile(common.CPPLanguage)
	if err != nil {
		t.Error(err)
	}
	exeFile, err = cpp.Compile(codeDir+"/hello.cpp", exeDir+"/hello_cpp")
	if err != nil {
		t.Error(err)
	}
	log.Println(exeFile)
}

func mkdirIfNotExist(path string, t *testing.T) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
}
