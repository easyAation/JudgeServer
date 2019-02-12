package dsl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"online_judge/talcity/scaffold/criteria/merr"
)

// Logical defines the logical for conditions
type Logical string

const (
	And Logical = "and"
	Or  Logical = "or"
)

var (
	supportedLogical = map[Logical]bool{
		And: true,
		Or:  true,
	}
)

// Validate check whether l is valid
func (l Logical) Validate() error {
	if !supportedLogical[l] {
		return merr.Wrap(nil, 0, "invalid logical: %s", l)
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type Logical
func (l Logical) MarshalJSON() ([]byte, error) {
	return []byte("\"" + string(l) + "\""), nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (l *Logical) UnmarshalJSON(value []byte) error {
	valStr, err := strconv.Unquote(string(value))
	if err != nil {
		return err
	}
	*l = Logical(valStr)
	return l.Validate()
}

// Filter represent for condition
type Filter struct {
	Field  string      `json:"field,omitempty"`
	Op     Operator    `json:"op,omitempty"`
	Target interface{} `json:"target,omitempty"`
}

// Validate check whether Filter is valid
func (c *Filter) Validate() error {
	if c.Field == "" {
		return merr.Wrap(nil, 0, "field is empty")
	}
	return c.Op.Validate()
}

// IsEmtpy check whether cond is empty
func (c *Filter) IsEmtpy() bool {
	return c.Field == "" && c.Op == "" && c.Target == nil
}

// ToSQL  convert struct to sql and args
// should make sure c.Validate() pass, otherwise, program maybe panic
func (c *Filter) ToSQL() (string, []interface{}, error) {
	iterm2Slice := func(target interface{}) ([]interface{}, error) {
		array := reflect.ValueOf(target)
		kind := array.Type().Kind()
		if kind != reflect.Array && kind != reflect.Slice {
			return nil, merr.Wrap(nil, 0, "target %v should be array")
		}
		length := array.Len()
		result := make([]interface{}, 0, length)
		for i := 0; i < length; i++ {
			result = append(result, array.Index(i).Interface())
		}
		return result, nil
	}
	switch c.Op {
	case Equal, Not, LessThan, LessEqual, GreaterThan, GreaterEqual, StartWith, EndWith, Like:
		switch c.Op {
		case StartWith:
			c.Target = fmt.Sprintf("%v%%", c.Target)
		case EndWith:
			c.Target = fmt.Sprintf("%%%v", c.Target)
		case Like:
			c.Target = fmt.Sprintf("%%%v%%", c.Target)
		}
		return fmt.Sprintf("%s %s ?", c.Field, c.Op.ToSQL()), []interface{}{c.Target}, nil
	case In, NotIn:
		targets, err := iterm2Slice(c.Target)
		if err != nil {
			return "", nil, err
		}
		if c.Op == In && len(targets) == 0 {
			return "0=1", nil, nil
		}
		if c.Op == NotIn && len(targets) == 0 {
			return "1=1", nil, nil
		}
		mask := strings.Repeat("?, ", len(targets))
		return fmt.Sprintf("%s %s (%s)", c.Field, c.Op.ToSQL(), mask[:len(mask)-2]), targets, nil
	default:
		return "", nil, merr.Wrap(nil, 0, "unsupported operation %s", c.Op)
	}
}
