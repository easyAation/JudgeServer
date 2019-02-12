package input

import (
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

type iterm struct {
	Num int    `json:"num" validate:"ne=0"`
	MSG string `json:"msg"`
}

func (i *iterm) Validate() error {
	if i.MSG == "Ops..." {
		return errors.New("shut up")
	}
	return nil
}

func TestVlidate(t *testing.T) {
	convey.Convey("Test Validate", t, func() {
		convey.Convey("test Validater", func() {
			i := &iterm{MSG: "Ops..."}
			err := Validate(i)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "shut up")
		})
		convey.Convey("test validate tag", func() {
			i := &iterm{}
			err := Validate(i)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldNotContainSubstring, "shut up")
			convey.So(err.Error(), convey.ShouldContainSubstring, "Field validation")
		})
	})
}
