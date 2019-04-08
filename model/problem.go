package model

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"online_judge/JudgeServer/common"
	"os"
	"strconv"
	"strings"
)

type DataFile struct {
	InputFile  string
	OutputFile string
}

type Problem struct {
	ID      int
	IOFiles []DataFile
}

var problems []Problem

func InitProblem() {
	problemsInfo, err := ioutil.ReadDir(common.Config.SandBox.ProblemDir)
	fmt.Println("problemdir: ", common.Config.SandBox.ProblemDir)
	if err != nil {
		panic(err)
	}
	for _, problemInfo := range problemsInfo {
		fmt.Println(problemInfo.Name(), " ", problemInfo.IsDir())
		problemDir := common.Config.SandBox.ProblemDir + string(os.PathSeparator) + problemInfo.Name()

		fmt.Println("path: ", problemDir)
		datasInfo, err := ioutil.ReadDir(problemDir)
		if err != nil {
			panic(err.Error())
			continue
		}
		fmt.Println("dataDir: ", datasInfo)
		var (
			problemItem Problem
		)
		problemItem.ID, err = strconv.Atoi(problemInfo.Name())
		if err != nil {
			continue
		}

		for _, dataInfo := range datasInfo {
			dataDir := problemDir + string(os.PathSeparator) + dataInfo.Name()
			datasFile, err := ioutil.ReadDir(dataDir)
			if err != nil {
				panic(err)
			}
			var dataFileItem DataFile
			for _, datafile := range datasFile {
				path := dataDir + string(os.PathSeparator) + datafile.Name()
				fmt.Println("datafile: ", path)
				if strings.Contains(datafile.Name(), "in") {
					dataFileItem.InputFile = path
				} else {
					dataFileItem.OutputFile = path
				}
			}
			problemItem.IOFiles = append(problemItem.IOFiles, dataFileItem)
		}
		problems = append(problems, problemItem)
	}
}

func GetProBlemByID(pid int) (*Problem, error) {
	for _, problem := range problems {
		if problem.ID == pid {
			return &problem, nil
		}
	}
	return nil, errors.Errorf("can't find problem of %d\n", pid)
}
