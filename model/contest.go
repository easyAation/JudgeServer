package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
)

const ContestTable = "contest"

type Contest struct {
	ID        int64     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Encrypt   int       `json:"encrypt" db:"encrypt"`
	StartAt   time.Time `json:"start_at" db:"start_at"`
	EndAt     time.Time `json:"end_at" db:"end_at"`
	CreatedAt time.Time `json:"create_at" db:"created_at"`
	UpdatedAt time.Time `json:"update_at" db:"updated_at"`
}

func (c *Contest) Valid() error {
	if c.Title == "" {
		return errors.Errorf("invalid title")
	}
	if c.StartAt.After(c.EndAt) {
		return errors.Errorf("The start time should not be greater than the end time.")
	}
	if c.StartAt.IsZero() {
		return errors.Errorf("invalid start time")
	}
	if c.EndAt.IsZero() {
		return errors.Errorf("invalid end time")
	}
	return nil
}

func AddContest(ctx context.Context, c Contest) (int64, error) {
	if err := c.Valid(); err != nil {
		return 0, err
	}
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return 0, err
	}
	result, err := sqlExec.Exec("INSERT INTO contest (title, encrypt, start_at, end_at) VALUES (?, ?,  ?, ?)",
		c.Title,
		c.Encrypt,
		c.StartAt, c.EndAt)
	if err != nil {
		return 0, errors.Wrap(err, "db error.")
	}
	return result.LastInsertId()
}

func GetContest(sqlExec *db.SqlExec, filters map[string]interface{}) ([]Contest, error) {
	placeHolder := make([]string, 0, len(filters))
	for key, value := range filters {
		// placeHolderValue = append(placeHolderValue, "?")
		placeHolder = append(placeHolder, fmt.Sprintf("%s='%v'", key, value))
	}
	sql := "SELECT * FROM " + ContestTable
	if len(placeHolder) != 0 {
		sql += " WHERE " + strings.Join(placeHolder, " AND ")
	}
	fmt.Println(sql)
	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, err
	}
	var contests []Contest
	for rows.Next() {
		var c Contest
		if err = rows.StructScan(&c); err != nil {
			return nil, errors.Wrap(err, "scan problem fail.")
		}
		contests = append(contests, c)
	}
	return contests, nil
}

func GetOneContest(sqlExec *db.SqlExec, filters map[string]interface{}) (*Contest, error) {
	cs, err := GetContest(sqlExec, filters)
	if err != nil {
		return nil, err
	}
	if len(cs) != 1 {
		return nil, errors.Errorf("expect one, but result is %d", len(cs))
	}
	return &cs[0], nil
}
