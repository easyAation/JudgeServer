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
	SubmitTable = "submit"
)

type Submit struct {
	ID        int64     `json:"id" db:"id"`
	PID       int       `json:"pid" db:"pid"`
	SubmitID  string    `json:"submit_id" db:"submit_id"`
	Code      string    `json:"code" db:"code"`
	Language  string    `json:"language" db:"language"`
	RunTime   int       `json:"run_time" db:"run_time"`
	Memory    int       `json:"memory" db:"memory"`
	Result    string    `json:"result" db:"result"`
	Author    string    `json:"author" db:"author"`
	CreatedAT time.Time `json:"created_at" db:"created_at"`
	UpdateAT  time.Time `json:"updated_at" db:"updated_at"`
}

func (submit *Submit) Valid() error {
	if submit.PID == 0 {
		return errors.Errorf("invalid pid")
	}
	if submit.SubmitID == "" {
		return errors.Errorf("invalid summit id")
	}
	if submit.Code == "" {
		return errors.Errorf("invalid code")
	}
	if submit.Language == "" {
		return errors.Errorf("language cannot be empty")
	}
	return nil
}

func AddSubmit(sqlExec *db.SqlExec, sm *Submit) (int64, error) {
	if err := sm.Valid(); err != nil {
		return 0, errors.Wrap(err, "invalid submit")
	}
	result, err := sqlExec.Exec("INSERT INTO submit (pid, submit_id, code, language, run_time, memory, result, author)"+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?)", sm.PID, sm.SubmitID, sm.Code, sm.Language, sm.RunTime, sm.Memory, sm.Result, sm.Author)
	if err != nil {
		return 0, errors.Wrap(err, "insert fail.")
	}
	return result.LastInsertId()
}

func UpdateSubmitBySID(sqlExec *db.SqlExec, sID string, values map[string]interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, errors.Errorf("invalid values. this is a empty values.")
	}
	placeHolder := make([]string, 0, len(values))
	for key, value := range values {
		if _, ok := value.(int); ok {
			placeHolder = append(placeHolder, fmt.Sprintf("%s=%v", key, value))
		} else {
			placeHolder = append(placeHolder, fmt.Sprintf("%s=\"%v\"", key, value))
		}
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE submit_id=\"%s\"", SubmitTable, strings.Join(placeHolder, " , "), sID)
	log.Println(sql)
	result, err := sqlExec.Exec(sql)
	if err != nil {
		return 0, errors.Wrap(err, "db error.")
	}
	return result.RowsAffected()
}

func GetSubmits(sqlExec *db.SqlExec, filters map[string]interface{}) ([]Submit, error) {
	placeHolder := make([]string, 0, len(filters))
	for key, value := range filters {
		placeHolder = append(placeHolder, fmt.Sprintf("%s=%v", key, value))
	}
	sql := "SELECT * FROM " + SubmitTable
	if len(placeHolder) != 0 {
		sql += " WHERE " + strings.Join(placeHolder, " AND ")
	}
	log.Println(sql)
	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, errors.Wrap(err, " ")
	}
	var sms []Submit
	for rows.Next() {
		var sm Submit
		if err := rows.StructScan(&sm); err != nil {
			return nil, errors.Wrap(err, "scan submit fail.")
		}
		sms = append(sms, sm)
	}
	return sms, nil
}

func GetOneSubmit(sqlExec *db.SqlExec, filters map[string]interface{}) (*Submit, error) {
	sms, err := GetSubmits(sqlExec, filters)
	if err != nil {
		return nil, err
	}
	if len(sms) != 1 {
		return nil, errors.Errorf("expect ont. but result is %d", len(sms))
	}
	return &sms[0], nil
}
