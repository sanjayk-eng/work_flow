package validator

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// Get returns singleton validator instance.
func Get() *validator.Validate {
	once.Do(func() {
		v := validator.New()

		// Use JSON tags in validation errors.
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

			if name == "-" {
				return ""
			}

			return name
		})

		registerCustomRules(v)

		instance = v
	})

	return instance
}

// Validate validates a struct and returns field errors.
func Validate(s any) map[string]string {
	err := Get().Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return map[string]string{
			"_": "invalid request",
		}
	}

	errs := make(map[string]string)

	for _, e := range ve {
		errs[e.Field()] = fieldMessage(e)
	}

	return errs
}

// Binder abstracts Gin Context for testing.
type Binder interface {
	ShouldBindJSON(any) error
}

// BindAndValidate binds JSON and validates it.
func BindAndValidate(c Binder, s any) (bool, map[string]string) {
	if err := c.ShouldBindJSON(s); err != nil {
		return false, map[string]string{
			"_": "invalid request payload",
		}
	}

	if errs := Validate(s); errs != nil {
		return false, errs
	}

	return true, nil
}

// fieldMessage converts validator tags to readable messages.
func fieldMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"

	case "email":
		return "must be a valid email address"

	case "min":
		return "must be at least " + e.Param() + " characters"

	case "max":
		return "must be at most " + e.Param() + " characters"

	case "uuid":
		return "must be a valid UUID"

	case "oneof":
		return "must be one of: " + strings.ReplaceAll(e.Param(), " ", ", ")

	case "url":
		return "must be a valid URL"

	case "strong_password":
		return "must contain uppercase, lowercase, digit and be at least 8 characters"

	default:
		return "invalid value"
	}
}

// registerCustomRules registers all custom validations.
func registerCustomRules(v *validator.Validate) {
	_ = v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		if len(password) < 8 {
			return false
		}

		var hasUpper bool
		var hasLower bool
		var hasDigit bool

		for _, r := range password {
			switch {
			case unicode.IsUpper(r):
				hasUpper = true
			case unicode.IsLower(r):
				hasLower = true
			case unicode.IsDigit(r):
				hasDigit = true
			}
		}

		return hasUpper && hasLower && hasDigit
	})
}
