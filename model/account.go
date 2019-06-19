package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
)

const AccountTable = "account"

type Account struct {
	ID         string `json:"id" db:"id"`
	Name       string `json:"name" db:"name"`
	Auth       string `json:"auth" db:"auth"` // 加密处理
	GitHupAddr string `json:"githup_addr" db:"githup_addr"`
	BlogAddr   string `json:"blog_addr" db:"blog_addr"`
}

func (ac *Account) Valid() error {
	if ac.ID == "" {
		return errors.Errorf("Account ID cannot be empty")
	}
	if ac.Name == "" {
		return errors.Errorf("Account name cannot be empty")
	}
	if ac.Auth == "" {
		return errors.Errorf("Account password cannot be empty")
	}
	return nil
}
func RegisterAccount(ctx context.Context, ac Account) error {
	if err := ac.Valid(); err != nil {
		return err
	}
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return err
	}
	result, err := sqlExec.Exec("INSERT INTO account (id, name, auth, githup_addr, blog_addr) "+
		"VALUES(?, ?, ?, ?, ?)", ac.ID, ac.Name, ac.Auth, ac.GitHupAddr, ac.BlogAddr)
	if err != nil {
		if strings.Contains(err.Error(), "PRIMARY") {
			err = errors.Errorf("Account number already exists")
		}
		return errors.Wrap(err, "")
	}
	_, err = result.LastInsertId()
	return err
}

func GetAccounts(ctx context.Context, filter map[string]interface{}) ([]Account, error) {
	sqlExec, err := db.GetSqlExec(ctx, "problem")
	if err != nil {
		return nil, err
	}
	placeHolder := make([]string, 0, len(filter))
	for key, value := range filter {
		placeHolder = append(placeHolder, fmt.Sprintf("%s='%v'", key, value))
	}
	sql := "SELECT * FROM " + AccountTable
	if len(placeHolder) != 0 {
		sql += " WHERE " + strings.Join(placeHolder, " AND ")
	}
	fmt.Println(sql)
	rows, err := sqlExec.Queryx(sql)
	if err != nil {
		return nil, err
	}
	var accounts []Account
	for rows.Next() {
		var account Account
		if err = rows.StructScan(&account); err != nil {
			return nil, errors.Wrap(err, "scan problem fail.")
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func GetOneAccount(ctx context.Context, filters map[string]interface{}) (*Account, error) {
	accounts, err := GetAccounts(ctx, filters)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, errors.Errorf("user not found.")
	}
	if len(accounts) != 1 {
		return nil, errors.Errorf("expect one, but result is %d", len(accounts))
	}
	return &accounts[0], nil
}
