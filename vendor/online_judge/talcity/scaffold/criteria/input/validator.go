package input

import (
	"sync"

	"online_judge/talcity/scaffold/util/validate"

	"gopkg.in/go-playground/validator.v9"
)

// TODO: implement sel-designed validator
var (
	checker     *validator.Validate
	checkerOnce sync.Once
)

func init() {
	lazyinit()
}

func lazyinit() {
	checkerOnce.Do(func() {
		checker = validator.New()
		checker.RegisterValidation("phone", ValidatePhone)
	})
}

func ValidatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return validate.ValidPhone(phone) == nil
}

func Validate(obj interface{}) error {
	fn, ok := obj.(Validater)
	if ok {
		if err := fn.Validate(); err != nil {
			return err
		}
	}
	return checker.Struct(obj)
}

// Validater any type who implement Validate() error
type Validater interface {
	Validate() error
}
