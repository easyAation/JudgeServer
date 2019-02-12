package trace_test

import (
	"context"
	"testing"

	"online_judge/talcity/scaffold/criteria/trace"

	"github.com/smartystreets/goconvey/convey"
)

func TestTrace(t *testing.T) {
	convey.Convey("test Trace", t, func() {
		t := trace.New()

		t.Add("part one", "process step %d", 1)
		t.Add("part one", "process step %d", 2)
		t.Add("part one", "process step %d", 3)
		t.Add("part two", "staging")
		t.Add("part three", "finish")

		keys, dic := t.Traces()
		convey.So(keys, convey.ShouldResemble, []string{"part one", "part two", "part three"})
		convey.So(dic, convey.ShouldResemble, map[string]interface{}{
			"part one":   []interface{}{"process step 1", "process step 2", "process step 3"},
			"part two":   "staging",
			"part three": "finish",
		})

		for _, k := range keys {
			convey.So(t.Value(k), convey.ShouldResemble, dic[k])
		}
		dic["not_exist_key"] = "blabal.."

		_, dic2 := t.Traces()
		convey.So(dic, convey.ShouldNotResemble, dic2)
	})
}

func TestTraceWithCtx(t *testing.T) {
	convey.Convey("test trace with context", t, func() {
		t := trace.New()
		ctx := trace.ToCtx(context.Background(), t)

		keys, dic := t.Traces()
		convey.So(keys, convey.ShouldResemble, []string{})
		convey.So(dic, convey.ShouldResemble, make(map[string]interface{}))

		t.Add("part one", "process step %d", 1)
		t.Add("part one", "process step %d", 2)
		t.Add("part one", "process step %d", 3)
		t.Add("part two", "staging")
		t.Add("part three", "finish")
		keys, dic = t.Traces()
		convey.So(keys, convey.ShouldResemble, []string{"part one", "part two", "part three"})
		convey.So(dic, convey.ShouldResemble, map[string]interface{}{
			"part one":   []interface{}{"process step 1", "process step 2", "process step 3"},
			"part two":   "staging",
			"part three": "finish",
		})

		t = trace.Get(ctx)
		keys, dic = t.Traces()
		convey.So(keys, convey.ShouldResemble, []string{"part one", "part two", "part three"})
		convey.So(dic, convey.ShouldResemble, map[string]interface{}{
			"part one":   []interface{}{"process step 1", "process step 2", "process step 3"},
			"part two":   "staging",
			"part three": "finish",
		})

		convey.Convey("test trace.Add", func() {
			tr := trace.New()
			ctx = trace.ToCtx(context.Background(), tr)
			trace.Add(ctx, "part one", "process step %d", 1)
			trace.Add(ctx, "part one", "process step %d", 2)
			trace.Add(ctx, "part one", "process step %d", 3)
			trace.Add(ctx, "part two", "staging")
			trace.Add(ctx, "part three", "finish")
			keys, dic = tr.Traces()
			convey.So(keys, convey.ShouldResemble, []string{"part one", "part two", "part three"})
			convey.So(dic, convey.ShouldResemble, map[string]interface{}{
				"part one":   []interface{}{"process step 1", "process step 2", "process step 3"},
				"part two":   "staging",
				"part three": "finish",
			})
		})
	})
}
