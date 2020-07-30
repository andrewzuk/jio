package jio

import (
    "errors"
	"regexp"
	"strings"
)

// String Generates a schema object that matches string data type
func String() *StringSchema {
	return &StringSchema{
		rules: make([]func(*Context), 0, 3),
	}
}

var _ Schema = new(StringSchema)

// StringSchema match string data type
type StringSchema struct {
	baseSchema

	required *bool
	rules    []func(*Context)
}

// SetPriority same as AnySchema.SetPriority
func (s *StringSchema) SetPriority(priority int) *StringSchema {
	s.priority = priority
	return s
}

// PrependTransform same as AnySchema.PrependTransform
func (s *StringSchema) PrependTransform(f func(*Context)) *StringSchema {
	s.rules = append([]func(*Context){f}, s.rules...)
	return s
}

// Transform same as AnySchema.Transform
func (s *StringSchema) Transform(f func(*Context)) *StringSchema {
	s.rules = append(s.rules, f)
	return s
}

// Custom adds a custom validation
func (s *StringSchema) Custom(name string, args ...interface{}) *StringSchema {
    return s.Transform(func(ctx *Context) {
        s.baseSchema.custom(ctx, name, args...)
    })
}

// Required same as AnySchema.Required
func (s *StringSchema) Required() *StringSchema {
	s.required = boolPtr(true)
	return s.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Abort(ErrorRequired(ctx))
		}
	})
}

// Optional same as AnySchema.Optional
func (s *StringSchema) Optional() *StringSchema {
	s.required = boolPtr(false)
	return s.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Skip()
		}
	})
}

// Default same as AnySchema.Default
func (s *StringSchema) Default(value string) *StringSchema {
	s.required = boolPtr(false)
	return s.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Value = value
		}
	})
}

// Set same as AnySchema.Set
func (s *StringSchema) Set(value string) *StringSchema {
	return s.Transform(func(ctx *Context) {
		ctx.Value = value
	})
}

// Equal same as AnySchema.Equal
func (s *StringSchema) Equal(value string) *StringSchema {
	return s.Check(func(ctxValue string) error {
		if value != ctxValue {
			return errors.New(ErrorMessageEqual(interface{}(value)))
		}
		return nil
	})
}

// When same as AnySchema.When
func (s *StringSchema) When(refPath string, condition interface{}, then Schema) *StringSchema {
	return s.Transform(func(ctx *Context) { s.whenEqual(ctx, refPath, condition, then) })
}

// Check use the provided function to validate the value of the key.
// Throws an error whenEqual the value is not string.
func (s *StringSchema) Check(f func(string) error) *StringSchema {
	return s.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(string)
		if !ok {
			ctx.Abort(ErrorTypeString(ctx))
			return
		}
		if err := f(ctxValue); err != nil {
			ctx.ErrorBag.Add(NewError(ctx, err.Error()))
		}
	})
}

// Valid same as AnySchema.Valid
func (s *StringSchema) Valid(values ...string) *StringSchema {
	return s.Check(func(ctxValue string) error {
		var isValid bool
		for _, v := range values {
			if v == ctxValue {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New(ErrorMessageStringOneOf(values))
		}
		return nil
	})
}

// Min check if the length of this string is greater than or equal to the provided length.
func (s *StringSchema) Min(min int) *StringSchema {
	return s.Check(func(ctxValue string) error {
		if len(ctxValue) < min {
			return errors.New(ErrorMessageStringLengthMin(min))
		}
		return nil
	})
}

// Max check if the length of this string is less than or equal to the provided length.
func (s *StringSchema) Max(max int) *StringSchema {
	return s.Check(func(ctxValue string) error {
		if len(ctxValue) > max {
			return errors.New(ErrorMessageStringLengthMin(max))
		}
		return nil
	})
}

// Length check if the length of this string is equal to the provided length.
func (s *StringSchema) Length(length int) *StringSchema {
	return s.Check(func(ctxValue string) error {
		if len(ctxValue) != length {
			return errors.New(ErrorMessageStringLengthEqual(length))
		}
		return nil
	})
}

// Regex check if the value is matched the regex.
func (s *StringSchema) Regex(regex string) *StringSchema {
	re := regexp.MustCompile(regex)
	return s.Check(func(ctxValue string) error {
		if !re.MatchString(ctxValue) {
			return errors.New(ErrorMessageMatchPattern(regex))
		}
		return nil
	})
}

// Alphanum check if the string value to only contain a-z, A-Z, and 0-9
func (s *StringSchema) Alphanum() *StringSchema {
	return s.Regex(`^[a-zA-Z0-9]+$`)
}

// Token check if the string value to only contain a-z, A-Z, 0-9, and underscore _
func (s *StringSchema) Token() *StringSchema {
	return s.Regex(`^\w+$`)
}

// Convert use the provided function to convert the value of the key.
// Throws an error whenEqual the value is not string.
func (s *StringSchema) Convert(f func(string) string) *StringSchema {
	return s.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(string)
		if !ok {
			ctx.Abort(ErrorTypeString(ctx))
			return
		}
		ctx.Value = f(ctxValue)
	})
}

// Lowercase convert the string value to lowercase.
func (s *StringSchema) Lowercase() *StringSchema {
	return s.Convert(strings.ToLower)

}

// Uppercase convert the string value to uppercase.
func (s *StringSchema) Uppercase() *StringSchema {
	return s.Convert(strings.ToUpper)
}

// Trim  emoves whitespace from both sides of the string value.
func (s *StringSchema) Trim() *StringSchema {
	return s.Convert(strings.TrimSpace)
}

// Validate same as AnySchema.Validate
func (s *StringSchema) Validate(ctx *Context) {
    if ctx.Value != nil {
        if _, ok := (ctx.Value).(string); !ok {
            ctx.Abort(ErrorTypeString(ctx))
            return
        }
    }
	if s.required == nil {
		s.Optional()
	}
	for _, rule := range s.rules {
		rule(ctx)
		if ctx.skip {
			return
		}
	}
}
