package compile

import (
	"log"
	"testing"

	"online_judge/JudgeServer/common"
)

func TestCompile(t *testing.T) {
	out, err := Compile("../test.c","test", common.CLanguage)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("%s\n", out)
	out, err = Compile("../cpp_test.cpp", "cpp_test", common.CPPLanguage)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("%s\n", out)
}
