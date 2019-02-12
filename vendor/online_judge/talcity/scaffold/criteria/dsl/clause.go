package dsl

import (
	"fmt"
	"strings"
	"sync"

	"online_judge/talcity/scaffold/criteria/merr"
)

const (
	CountSQLUnit = "COUNT(1)"
)

var (
	maxClauseDepth uint = 3
	once           sync.Once
)

// SetMaxClauseDepth global set the max clause depth, default is 3
// if clause max depth greater than the max depth, clause.Validate() will return
// err: clauses too deep
func SetMaxClauseDepth(depth uint) error {
	if depth < 1 {
		return merr.Wrap(nil, 0, "depth must greater than 0")
	}
	once.Do(func() { maxClauseDepth = depth })
	return nil
}

// Clause represent for bool query stuct,
// Filter (Field, Op, Target) and (Logical, Clauses) are mutually exclusive
// a deomo for Clause:
// {
// 	"logical": "and",
// 	"clauses": [
// 		{"field": "col1", "op": "like", "target": "x1"},
// 		{"field": "col2", "op": "not_eq", "target": "x2"},
// 		{
// 			"logical": "or",
// 			"clauses": [
// 				{"field": "col3", "op": "gte", "target": "123"},
// 				{"field": "col4", "op": "lt", "target": "2"},
// 				{"field": "col5", "op": "in", "target": [1, 3, 5]}
// 			]
// 		}
// 	]
// }
type Clause struct {
	Filter
	Logical Logical  `json:"logical,omitempty"`
	Clauses []Clause `json:"clauses,omitempty"`
}

// RenameField replace filter field from old to new
func (c *Clause) RenameField(old, new string) {
	if c.Field == old {
		c.Field = new
	}
	for _, child := range c.Clauses {
		child.RenameField(old, new)
	}
}

// CheckDepth check whether Clause's max deep equal or greater than depth
func (c *Clause) CheckDepth(depth uint) bool {
	if c == nil {
		return depth == 0
	}
	if depth <= 1 {
		return true
	}
	for _, child := range c.Clauses {
		if child.CheckDepth(depth - 1) {
			return true
		}
	}
	return false
}

// Validate check whether Clause is valid
func (c *Clause) Validate() (err error) {
	if !c.Filter.IsEmtpy() {
		// children must empty
		if c.Logical != "" || len(c.Clauses) > 0 {
			return merr.Wrap(nil, 0, "conditon (field, op, target) and (logical, clauses) are mutually exclusive")
		}
		return c.Filter.Validate()
	}
	// cond is empty, children must valid
	if err = c.Logical.Validate(); err != nil {
		return err
	}
	if len(c.Clauses) == 0 {
		return merr.Wrap(nil, 0, "logical and clauses should both emtpty or both valued")
	}
	for _, iterm := range c.Clauses {
		if err = iterm.Validate(); err != nil {
			return err
		}
	}
	// deep less than maxClauseDepth
	if c.CheckDepth(maxClauseDepth + 1) {
		return merr.Wrap(nil, 0, "clauses too deep")
	}
	return nil
}

// ToSQL convert c to sql-format conditon check and args
// eg: "col1 like %x1% and col2 != x2 and (col3 >= 123 orcol4 < 2 or col5 in (1, 3, 5))"
// select f1, f2 from table where %s
// should make sure c.Validate() pass, otherwise, program maybe panic
func (c *Clause) ToSQL() (string, []interface{}, error) {
	if !c.Filter.IsEmtpy() {
		return c.Filter.ToSQL()
	}

	conds, args := make([]string, 0, 10), make([]interface{}, 0, 10)

	for _, child := range c.Clauses {
		cond, vals, err := child.ToSQL()
		if err != nil {
			return "", nil, err
		}
		conds = append(conds, cond)
		args = append(args, vals...)
	}
	if len(conds) == 1 {
		return conds[0], args, nil
	}
	for i, v := range conds {
		conds[i] = "(" + v + ")"
	}

	return strings.Join(conds, " "+string(c.Logical)+" "), args, nil
}

type SQLSearchUnity struct {
	Page     *int       `json:"page,omitempty"`
	Size     *int       `json:"size,omitempty"`
	Clause   *Clause    `json:"matcher,omitempty"`
	OrderBys [][]string `json:"order_bys"`
}

func (su *SQLSearchUnity) Validate() (err error) {
	if su.Size != nil && *su.Size < 1 {
		return merr.Wrap(nil, 0, "size should omit or greater than 0")
	}
	if su.Page != nil {
		if *su.Page < 1 {
			return merr.Wrap(nil, 0, "page should omit or greater than 0")
		}
		if su.Size == nil {
			return merr.Wrap(nil, 0, "size should not empty when page hold value")
		}
	}
	if su.Clause != nil {
		if err = su.Clause.Validate(); err != nil {
			return err
		}
	}
	for _, order := range su.OrderBys {
		if len(order) != 2 {
			return merr.Wrap(nil, 0, "order by should contains two iterm")
		}
		if order[1] != "DESC" && order[1] != "ASC" {
			return merr.Wrap(nil, 0, "order by sort method not valid. only support DESC and ASC")
		}
	}
	return nil
}

// BuildSearchSQL build sql search code and args
// should make sure su.Validate() pass, otherwise, program maybe panic
func BuildSearchSQL(table string, fields []string, su SQLSearchUnity) (string, []interface{}, error) {
	SQL, args, err := BuildCountSQL(table, su.Clause)
	if err != nil {
		return "", nil, err
	}
	SQL = strings.Replace(SQL, CountSQLUnit, strings.Join(fields, ", "), -1)

	if len(su.OrderBys) > 0 {
		sorts := make([]string, 0, len(su.OrderBys))
		for _, order := range su.OrderBys {
			sorts = append(sorts, order[0]+" "+order[1])
		}
		SQL = SQL + " ORDER BY " + strings.Join(sorts, " , ")
	}
	if su.Size != nil {
		var from int
		if su.Page != nil {
			from = (*su.Page - 1) * (*su.Size)
		}
		SQL = SQL + " LIMIT ?, ?"
		args = append(args, from, *su.Size)
	}

	return SQL, args, nil
}

// BuildCountSQL build sql search code and args
// should make sure matcher.Validate() pass, otherwise, program maybe panic
func BuildCountSQL(table string, matcher *Clause) (string, []interface{}, error) {
	var (
		conditions string
		args       []interface{}
		err        error
	)
	if matcher == nil {
		conditions = "1=1"
	} else {
		conditions, args, err = matcher.ToSQL()
		if err != nil {
			return "", nil, err
		}
	}
	SQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", CountSQLUnit, table, conditions)

	return SQL, args, nil
}
