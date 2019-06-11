package model

import (
	"fmt"
	"strings"

	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
)

const ContestSubmitTable = "contest_submit"

type ContestSubmit struct {
	Submit
	CID int64 `json:"cid" db:"cid"`
}

func (c *ContestSubmit) Valid() error {
	if c.CID == 0 {
		return errors.Errorf("invalid contest id")
	}
	return c.Submit.Valid()
}

func AddContestSubmit(sqlExec *db.SqlExec, cs ContestSubmit) (int64, error) {
	if err := cs.Valid(); err != nil {
		return 0, errors.Wrap(err, "invalid submit")
	}
	result, err := sqlExec.Exec("INSERT INTO contest_submit (pid, uid, cid, submit_id, code, language, run_time, "+
		"memory, result)"+
		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		cs.PID, cs.UID, cs.CID, cs.SubmitID, cs.Code, cs.Language, cs.RunTime, cs.Memory, cs.Result)
	if err != nil {
		return 0, errors.Wrap(err, "insert fail.")
	}
	return result.LastInsertId()
}

func GetContestSubmit(sqlExec *db.SqlExec, filters map[string]interface{}) ([]ContestSubmit, error) {
	placeHolder := make([]string, 0, len(filters))
	for key, value := range filters {
		// placeHolderValue = append(placeHolderValue, "?")
		placeHolder = append(placeHolder, fmt.Sprintf("%s='%v'", key, value))
	}
	sql := "SELECT * FROM " + ContestSubmitTable
	if len(placeHolder) != 0 {
		sql += " WHERE " + strings.Join(placeHolder, " AND ")
	}
	fmt.Println(sql)
	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, err
	}
	var css []ContestSubmit
	for rows.Next() {
		var cs ContestSubmit
		if err = rows.StructScan(&cs); err != nil {
			return nil, errors.Wrap(err, "")
		}
		css = append(css, cs)
	}
	return css, nil
}
