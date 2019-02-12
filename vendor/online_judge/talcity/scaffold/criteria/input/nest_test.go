package input_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"online_judge/talcity/scaffold/criteria/input"
	"online_judge/talcity/scaffold/criteria/route"
)

func TestNest(t *testing.T) {
	convey.Convey("Test nest", t, func(c convey.C) {
		studentBodyJSON := `{
			"name": "abc",
			"age": 15,
			"questions": [
				{"id": 0, "closed": false},
				{"id": 1, "closed": true}
			]
		}`
		type student struct {
			Name      *string `json:"name"`
			Age       int     `json:"age"`
			Questions []struct {
				ID     int  `json:"id"`
				Closed bool `json:"closed"`
			} `json:"questions"`
		}
		s := httptest.NewServer(route.BuildHandler(nil, &route.ModuleRoute{
			Routes: []*route.Route{
				route.NewRoute("/accounts/:id", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
					p := input.NewParam(r)

					nest := struct {
						Token   *string  `header:"token"`
						ID2     *string  `var:"id"`
						Page    *int     `query:"page"`
						ID      string   `query:"id"`
						Student *student `body:"body"`
						Score   float64  `query:"score" validate:"gte=0"`
					}{}

					c.So(p.Nest(&nest).Error(), convey.ShouldBeNil)
					var expectedStudent student
					c.So(json.Unmarshal([]byte(studentBodyJSON), &expectedStudent), convey.ShouldBeNil)
					c.So(*nest.Student, convey.ShouldResemble, expectedStudent)
					c.So(nest.Score, convey.ShouldEqual, 84.5)
					c.So(*nest.ID2, convey.ShouldEqual, nest.ID)
					w.WriteHeader(http.StatusOK)
				}, nil),
				route.NewRoute("/jobs", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
					p2 := input.NewParam(r)

					nest := struct {
						Token   *string  `header:"token" validate:"len=22"`
						Page    *int     `query:"page"`
						ID      string   `query:"id"`
						Student *student `body:"body"`
					}{}
					c.So(p2.Nest(&nest).Error().Error(), convey.ShouldContainSubstring, `'Token' failed`)
					w.WriteHeader(http.StatusOK)
				}, nil),
				route.NewRoute("/default/test", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
					p := input.NewParam(r)

					nest := struct {
						Page  *int    `query:"page"`
						Score float64 `query:"score" validate:"gte=0"`
					}{}

					c.So(p.Nest(&nest).Error(), convey.ShouldBeNil)
					c.So(nest.Score, convey.ShouldEqual, 0)
					c.So(nest.Page, convey.ShouldBeNil)
					w.WriteHeader(http.StatusOK)
				}, nil),
			},
		}))
		defer s.Close()

		c.Convey("Test marshal nest value", func() {
			req, err := http.NewRequest(http.MethodGet, s.URL+"/accounts/aaaa?id=aaaa&page=10&score=84.5", bytes.NewBuffer([]byte(studentBodyJSON)))
			c.So(err, convey.ShouldBeNil)
			req.Header.Add("token", "def")

			resp, err := s.Client().Do(req)
			c.So(err, convey.ShouldBeNil)
			c.So(resp.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		c.Convey("Test marshal nest value with default value", func() {
			req, err := http.NewRequest(http.MethodGet, s.URL+"/default/test", bytes.NewBuffer([]byte(studentBodyJSON)))
			c.So(err, convey.ShouldBeNil)

			resp, err := s.Client().Do(req)
			c.So(err, convey.ShouldBeNil)
			c.So(resp.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
		c.Convey("Test validator failed", func() {
			req, err := http.NewRequest(http.MethodGet, s.URL+"/jobs?id=aaaa&page=10", bytes.NewBuffer([]byte(studentBodyJSON)))
			c.So(err, convey.ShouldBeNil)
			req.Header.Add("token", "def")

			resp, err := s.Client().Do(req)
			c.So(err, convey.ShouldBeNil)
			c.So(resp.StatusCode, convey.ShouldEqual, http.StatusOK)
		})
	})
}
