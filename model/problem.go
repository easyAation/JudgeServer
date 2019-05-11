package model

import (
	"fmt"
	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
	"log"
	"strings"
	"time"
)

const (
	ProblemTable = "problem"
)

type Problem struct {
	ID             int       `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Author         string    `json:"author" db:"author"`
	Status         string    `json:"status,omitempty" db:"status,omitempty"`
	Difficulty     string    `json:"difficulty" db:"difficulty"`
	CaseDataInput  string    `json:"case_data_input" db:"case_data_input"`
	CaseDataOutput string    `json:"case_data_output" db:"case_data_output"`
	Description    string    `json:"description" db:"description"`
	InputDes       string    `json:"input_des" db:"input_des"`
	OutputDes      string    `json:"output_des" db:"output_des"`
	Hint           string    `json:"hint" db:"hint"`
	TimeLimit      int64     `json:"time_limit" db:"time_limit"`
	MemoryLimit    int64     `json:"memory_limit" db:"memory_limit"`
	AuthorCode     string    `json:"author_code" db:"author_code"`
	CreatedTime    time.Time `json:"create_time" db:"created_time"`
	UpdatedTime    time.Time `json:"update_time" db:"updated_time"`
}

func (pro *Problem) Valid() error {
	if pro.ID == 0 {
		return errors.Errorf("invalid id")
	}
	if pro.Name == "" {
		return errors.Errorf("invalid name")
	}
	if pro.Author == "" {
		return errors.Errorf("invalid author")
	}
	if pro.TimeLimit == 0 {
		return errors.Errorf("invalid tile limit")
	}
	if pro.MemoryLimit == 0 {
		return errors.Errorf("invalid memory limit")
	}
	return nil
}

func AddProblem(sqlExec *db.SqlExec, pro Problem) (int64, error) {
	if err := pro.Valid(); err != nil {
		return 0, err
	}
	result, err := sqlExec.Exec("INSERT INTO problem (id, name, author, status, difficulty, case_data_input, case_data_output, description, input_des, output_des, hint, time_limit,memory_limit) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", pro.ID, pro.Name, pro.Author, pro.Status, pro.Difficulty, pro.CaseDataInput, pro.CaseDataOutput, pro.Description, pro.InputDes, pro.OutputDes, pro.Hint, pro.TimeLimit, pro.MemoryLimit)
	if err != nil {
		return 0, errors.Wrap(err, "db error.")
	}

	return result.LastInsertId()
}

func GetProblem(sqlExec *db.SqlExec, filters map[string]interface{}) ([]Problem, error) {
	placeHolder := make([]string, 0, len(filters))
	for key, value := range filters {
		// placeHolderValue = append(placeHolderValue, "?")
		placeHolder = append(placeHolder, fmt.Sprintf("%s=%v", key, value))
	}
	sql := "SELECT * FROM " + ProblemTable
	if len(placeHolder) != 0 {
		sql += " WHERE " + strings.Join(placeHolder, " AND ")
	}
	fmt.Println(sql)
	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, err
	}
	var problems []Problem
	for rows.Next() {
		var pro Problem
		if err = rows.StructScan(&pro); err != nil {
			return nil, errors.Wrap(err, "scan problem fail.")
		}
		problems = append(problems, pro)
	}
	return problems, nil
}

func GetOneProblem(sqlExec *db.SqlExec, filters map[string]interface{}) (*Problem, error) {
	problems, err := GetProblem(sqlExec, filters)
	if err != nil {
		return nil, err
	}
	if len(problems) != 1 {
		return nil, errors.Errorf("expect one, but result is %d", len(problems))
	}
	return &problems[0], nil
}

func UpdateProblem(sqlExec *db.SqlExec, id int64, values map[string]interface{}) (int64, error) {
	placeHolder := make([]string, 0, len(values))
	for key, value := range values {
		if _, ok := value.(int); ok {
			placeHolder = append(placeHolder, fmt.Sprintf("%s=%v", key, value))
		} else {
			placeHolder = append(placeHolder, fmt.Sprintf("%s=\"%v\"", key, value))
		}
	}
	sql := fmt.Sprintf("UPDATE problem set %s where id = %d", strings.Join(placeHolder, " , "), id)
	log.Println(sql)
	result, err := sqlExec.Exec(sql)
	if err != nil {
		return 0, errors.Wrap(err, "sql exec fail.")
	}
	return result.RowsAffected()
}
