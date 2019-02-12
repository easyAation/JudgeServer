package dsl

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCompare(t *testing.T) {
	convey.Convey("Test Clause", t, func(c convey.C) {
		rawJSON := `{
			"logical": "and",
			"clauses": [
				{"field": "col1", "op": "like", "target": "x1"},
				{"field": "col2", "op": "not_eq", "target": "x2"},
				{
					"logical": "or",
					"clauses": [
						{"field": "col3", "op": "gte", "target": "123"},
						{"field": "col4", "op": "lt", "target": "2"},
						{"field": "col5", "op": "in", "target": [1, 3, 5]}
					]
				}
			]
		}`
		var leaf Clause
		err := json.Unmarshal([]byte(rawJSON), &leaf)
		convey.So(err, convey.ShouldBeNil)
		b, err := json.Marshal(leaf)
		convey.So(err, convey.ShouldBeNil)

		// test json marshaler
		rawMap, newMap := make(map[string]interface{}), make(map[string]interface{})
		err1, err2 := json.Unmarshal([]byte(rawJSON), &rawMap), json.Unmarshal(b, &newMap)
		convey.So(err1, convey.ShouldBeNil)
		convey.So(err2, convey.ShouldBeNil)
		convey.So(rawMap, convey.ShouldResemble, rawMap)
		// test deep check
		convey.So(leaf.CheckDepth(1), convey.ShouldBeTrue)
		convey.So(leaf.CheckDepth(2), convey.ShouldBeTrue)
		convey.So(leaf.CheckDepth(3), convey.ShouldBeTrue)
		convey.So(leaf.Validate(), convey.ShouldBeNil)
		convey.So(leaf.CheckDepth(4), convey.ShouldBeFalse)
		// test set max depth, only the first set work
		SetMaxClauseDepth(3)
		convey.So(leaf.Validate(), convey.ShouldBeNil)
		convey.So(leaf.CheckDepth(4), convey.ShouldBeFalse)
		SetMaxClauseDepth(4)
		convey.So(leaf.CheckDepth(4), convey.ShouldBeFalse)
		convey.So(leaf.CheckDepth(3), convey.ShouldBeTrue)
		convey.So(leaf.Validate(), convey.ShouldBeNil)

		// test build sql
		condSQL, args, err := leaf.ToSQL()
		convey.So(err, convey.ShouldBeNil)
		convey.So(condSQL, convey.ShouldEqual, "(col1 LIKE ?) and (col2 != ?) and ((col3 >= ?) or (col4 < ?) or (col5 IN (?, ?, ?)))")
		convey.So(interfaceArrEqual(args, []interface{}{"%x1%", "x2", "123", "2", 1, 3, 5}), convey.ShouldBeTrue)
	})
}
func TestSQLSearchUnity(t *testing.T) {
	convey.Convey("Test SQLSearchUnity", t, func(c convey.C) {
		rawJSON := `{
			"page": 1,
			"size": 20,
			"matcher": {
				"logical": "and",
				"clauses": [
					{"field": "col1", "op": "like", "target": "x1"},
					{"field": "col2", "op": "not_eq", "target": "x2"},
					{
						"logical": "or",
						"clauses": [
							{"field": "col3", "op": "gte", "target": "123"},
							{"field": "col4", "op": "lt", "target": "2"},
							{"field": "col5", "op": "in", "target": [1, 3, 5]}
						]
					}
				]
			},
			"ordery_bys": [
				["col1", "asc"],
				["col2", "desc"]
			]
		}`
		var searchUnit SQLSearchUnity
		err := json.Unmarshal([]byte(rawJSON), &searchUnit)
		convey.So(err, convey.ShouldBeNil)
		b, err := json.Marshal(searchUnit)
		convey.So(err, convey.ShouldBeNil)

		rawMap, newMap := make(map[string]interface{}), make(map[string]interface{})
		err1, err2 := json.Unmarshal([]byte(rawJSON), &rawMap), json.Unmarshal(b, &newMap)
		convey.So(err1, convey.ShouldBeNil)
		convey.So(err2, convey.ShouldBeNil)
		convey.So(rawMap, convey.ShouldResemble, rawMap)

		countSQL, args, err := BuildCountSQL("table1", searchUnit.Clause)
		convey.So(err, convey.ShouldBeNil)
		convey.So(countSQL, convey.ShouldEqual, "SELECT COUNT(1) FROM table1 WHERE (col1 LIKE ?) and (col2 != ?) and ((col3 >= ?) or (col4 < ?) or (col5 IN (?, ?, ?)))")
		convey.So(interfaceArrEqual(args, []interface{}{"%x1%", "x2", "123", "2", 1, 3, 5}), convey.ShouldBeTrue)

		searchSQL, args, err := BuildSearchSQL("table1", []string{"col1", "col2", "col3"}, searchUnit)
		convey.So(err, convey.ShouldBeNil)
		convey.So(searchSQL, convey.ShouldEqual, "SELECT col1, col2, col3 FROM table1 WHERE (col1 LIKE ?) and (col2 != ?) and ((col3 >= ?) or (col4 < ?) or (col5 IN (?, ?, ?))) LIMIT ?, ?")
		convey.So(interfaceArrEqual(args, []interface{}{"%x1%", "x2", "123", "2", 1, 3, 5, 0, 20}), convey.ShouldBeTrue)
	})
}

func interfaceArrEqual(expect, target []interface{}) bool {
	if len(expect) != len(target) {
		return false
	}
	b1, e1 := json.Marshal(expect)
	b2, e2 := json.Marshal(target)
	if e1 != nil {
		panic(e1)
	}
	if e2 != nil {
		panic(e2)
	}
	return reflect.DeepEqual(b1, b2)
}
