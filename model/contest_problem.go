package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
)

const ContestProblemTable = "contest_problem"

type ContestProblem struct {
	ID          int64     `json:"id" db:"id"`
	PID         int64     `json:"pid" db:"pid"`
	CID         int64     `json:"cid" db:"cid"`
	Title       string    `json:"title"`
	Position    int       `json:"position" db:"position"`
	SubmitCount int       `json:"submit_count" db:"submit_count"`
	SolveCount  int       `json:"solve_count" db:"solve_count"`
	CreatedAT   time.Time `json:"create_at" db:"created_at"`
	UpdatedAT   time.Time `json:"update_at" db:"updated_at"`
}

func (cp *ContestProblem) Valid() error {
	if cp.PID == 0 {
		return errors.Errorf("invalid PID")
	}
	if cp.CID == 0 {
		return errors.Errorf("invalid CID")
	}
	return nil
}

func AddContestProblem(ctx context.Context, cp ContestProblem) (int64, error) {
	if err := cp.Valid(); err != nil {
		return 0, err
	}
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return 0, errors.Errorf("db error.")
	}
	result, err := sqlExec.Exec("INSERT INTO contest_problem (pid, cid, position, submit_count, "+
		"solve_count) VALUES (?, ?, , ?, ?, ?)", cp.PID, cp.CID, cp.Position, cp.SolveCount, cp.SolveCount)
	if err != nil {
		return 0, errors.Errorf("db error")
	}
	return result.LastInsertId()
}

func GetContestProblems(sqlExec *db.SqlExec, filters map[string]interface{}) ([]ContestProblem, error) {
	placeHolder := make([]string, 0, len(filters))
	for key, value := range filters {
		// placeHolderValue = append(placeHolderValue, "?")
		placeHolder = append(placeHolder, fmt.Sprintf("%s='%v'", key, value))
	}
	sql := "SELECT * FROM " + ContestProblemTable
	if len(placeHolder) != 0 {
		sql += " WHERE " + strings.Join(placeHolder, " AND ")
	}
	fmt.Println(sql)
	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, err
	}
	var cps []ContestProblem
	for rows.Next() {
		var cp ContestProblem
		if err = rows.StructScan(&cp); err != nil {
			return nil, errors.Wrap(err, "")
		}
		cps = append(cps, cp)
	}
	return cps, nil
}

// func UpdateSubmit(ctx context.Context, cid, pid int) error {
//         sqlExec, err := db.GetSqlExec(ctx, "problem")
//         if err != nil {
//                 return errors.Errorf("db error.")
//         }
//         result, err := sqlExec.Exec("UPDATE contest_problem set submit")
// }
