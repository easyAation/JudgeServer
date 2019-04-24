package model

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
)

type DataFile struct {
	InputFile  string
	OutputFile string
}

type ProblemData struct {
	ID           int    `json:"id" db:"id"`
	PID          int    `json:"pid" db:"pid"`
	InputFile    string `json:"input_file" db:"input_file"`
	OutputFile   string `json:"output_file" db:"output_file"`
	MD5          string `json:"md5" db:"md5"`
	MD5TrimSpace string `json:"md5_trim_space" db:"md5_trim_space"`
}

func (proData *ProblemData) CalculMD5() {

}
func AddProblemDatas(sqlExec *db.SqlExec, proDatas []ProblemData) (int64, error) {
	fn := func() (int64, error) {
		var rows sql.Result
		var err error
		tx := sqlExec.MustBegin()
		for _, proData := range proDatas {
			rows, err = tx.NamedExec("INSERT INTO problem_data (id, pid, input_file, output_file, md5,"+
				"md5_trim_space) VALUES (:id, :pid, :input_file, :output_file, :md5, :md5_trim_space)", &proData)
			if err != nil {
				return 0, errors.Wrap(err, "internal error.")
			}
		}
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return rows.LastInsertId()
	}
	return fn()
}

func GetProblemData(sqlExec *db.SqlExec, filter map[string]interface{}) ([]ProblemData, error) {
	placeHolder := make([]string, 0, len(filter))
	for k, v := range filter {
		placeHolder = append(placeHolder, fmt.Sprintf("%s=%v ", k, v))
	}
	sql := "select * from problem_data where " + strings.Join(placeHolder, "and")
	fmt.Println(sql)

	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, errors.Wrap(err, "query error.")
	}
	var prodatas []ProblemData
	for rows.Next() {
		var prodata ProblemData
		err := rows.StructScan(&prodata)
		if err != nil {
			log.Print(err)
			continue
		}
		prodatas = append(prodatas, prodata)
	}
	return prodatas, nil
}
