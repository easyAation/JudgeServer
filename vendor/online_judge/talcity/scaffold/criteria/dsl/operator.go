package dsl

import (
	"strconv"

	"online_judge/talcity/scaffold/criteria/merr"
)

// Operator defines the operators for comparison
type Operator string

const (
	// Operation
	Equal        Operator = "eq"
	Not          Operator = "not_eq"
	LessThan     Operator = "lt"
	LessEqual    Operator = "lte"
	GreaterThan  Operator = "gt"
	GreaterEqual Operator = "gte"
	In           Operator = "in"
	NotIn        Operator = "not_in"
	StartWith    Operator = "starts_with"
	EndWith      Operator = "ends_with"
	Like         Operator = "like"
)

var (
	supportedOperation = map[Operator]bool{
		Equal:        true,
		Not:          true,
		LessThan:     true,
		LessEqual:    true,
		GreaterThan:  true,
		GreaterEqual: true,
		In:           true,
		NotIn:        true,
		StartWith:    true,
		EndWith:      true,
		Like:         true,
	}
)

// Validate check whether c is valid
func (c Operator) Validate() error {
	if !supportedOperation[c] {
		return merr.Wrap(nil, 0, "invalid operator: %s ", c)
	}
	return nil
}

func (c Operator) ToSQL() string {
	switch c {
	case Equal:
		return "="
	case Not:
		return "!="
	case LessThan:
		return "<"
	case LessEqual:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterEqual:
		return ">="
	case In:
		return "IN"
	case NotIn:
		return "NOT IN"
	case StartWith, EndWith, Like:
		return "LIKE"
	default:
		return string(c)
	}
}

// MarshalJSON implements the json.Marshaller interface for type Operator
func (c Operator) MarshalJSON() ([]byte, error) {
	return []byte("\"" + string(c) + "\""), nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (c *Operator) UnmarshalJSON(value []byte) error {
	valStr, err := strconv.Unquote(string(value))
	if err != nil {
		return err
	}
	*c = Operator(valStr)
	return c.Validate()
}
