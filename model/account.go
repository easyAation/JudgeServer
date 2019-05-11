package model

import (
	"github.com/easyAation/scaffold/db"
	"github.com/pkg/errors"
)

type Account struct {
	ID         string `json:"id" db:"id"`
	Name       string `json:"name" db:"name"`
	Password   string `json:"password" db:"password"` // 加密处理
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
	if ac.Password == "" {
		return errors.Errorf("Account password cannot be empty")
	}
	return nil
}
func RegisterAccout(sqlExec *db.SqlExec, ac Account) error {
	if err := ac.Valid(); err != nil {
		return err
	}
	result, err := sqlExec.Exec("INSERT INTO account (id, name, password, githup_addr, blog_addr) "+
		"VALUES(?, ?, ?, ?, ?", ac.ID, ac.Name, ac.Password, ac.GitHupAddr, ac.BlogAddr)
	if err != nil {
		return errors.Wrap(err, "db error.")
	}
	_, err = result.LastInsertId()
	return err
}
