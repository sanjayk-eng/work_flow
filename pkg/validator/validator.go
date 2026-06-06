package validator

import (
	"errors"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// Get returns the shared validator instance (singleton).
func Get() *validator.Validate {
	once.Do(func() {
		instance = validator.New(validator.WithRequiredStructFields())
		registerCustomRules(instance)
	})
	return instance
}

// Validate binds and validates a struct. Returns a map of field → message on failure.
func Validate(s any) map[string]string {
	err := Get().Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return map[string]string{"_": "invalid request"}
	}

	errs := make(map[string]string, len(ve))
	for _, e := range ve {
		field := toSnakeCase(e.Field())
		errs[field] = fieldMessage(e)
	}
	return errs
}

// BindAndValidate binds JSON from a gin context and validates the struct in one call.
// Returns (true, nil) on success.
// Returns (false, map) on validation failure — map is ready to send as JSON errors.
// Returns (false, map{"_": "..."}) on JSON parse failure.
func BindAndValidate(c interface {
	ShouldBindJSON(any) error
}, s any) (ok bool, errs map[string]string) {
	type binder interface {
		ShouldBindJSON(any) error
	}
	if err := c.ShouldBindJSON(s); err != nil {
		return false, map[string]string{"_": "invalid JSON body"}
	}
	if errs := Validate(s); errs != nil {
		return false, errs
	}
	return true, nil
}

// fieldMessage returns a human-readable message for a validation error.
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
	case "uuid4":
		return "must be a valid UUID"
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(e.Param(), " ", ", ")
	case "url":
		return "must be a valid URL"
	case "strong_password":
		return "must be at least 8 characters with uppercase, lowercase, and a number"
	default:
		return "is invalid"
	}
}

// registerCustomRules adds project-specific validation rules.
func registerCustomRules(v *validator.Validate) {
	// strong_password: min 8 chars, at least one upper, one lower, one digit
	_ = v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) < 8 {
			return false
		}
		var hasUpper, hasLower, hasDigit bool
		for _, c := range val {
			switch {
			case c >= 'A' && c <= 'Z':
				hasUpper = true
			case c >= 'a' && c <= 'z':
				hasLower = true
			case c >= '0' && c <= '9':
				hasDigit = true
			}
		}
		return hasUpper && hasLower && hasDigit
	})
}

// toSnakeCase converts "FirstName" → "first_name" for consistent JSON field names in errors.
func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(r | 0x20) // to lower
	}
	return b.String()
}
