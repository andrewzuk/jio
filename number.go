package jio

import (
    "errors"
    "math"
    "strconv"
)

// Number Generates a schema object that matches number data type
func Number() *NumberSchema {
	return &NumberSchema{
		rules: make([]func(*Context), 0, 3),
	}
}

var _ Schema = new(NumberSchema)

// NumberSchema match number data type
type NumberSchema struct {
	baseSchema

	required *bool
	rules    []func(*Context)
}

// SetPriority same as AnySchema.SetPriority
func (n *NumberSchema) SetPriority(priority int) *NumberSchema {
	n.priority = priority
	return n
}

// PrependTransform same as AnySchema.PrependTransform
func (n *NumberSchema) PrependTransform(f func(*Context)) *NumberSchema {
	n.rules = append([]func(*Context){f}, n.rules...)
	return n
}

// Transform same as AnySchema.Transform
func (n *NumberSchema) Transform(f func(*Context)) *NumberSchema {
	n.rules = append(n.rules, f)
	return n
}

// Custom adds a custom validation
func (n *NumberSchema) Custom(name string, args ...interface{}) *NumberSchema {
    return n.Transform(func(ctx *Context) {
        n.baseSchema.custom(ctx, name, args...)
    })
}

// Required same as AnySchema.Required
func (n *NumberSchema) Required() *NumberSchema {
	n.required = boolPtr(true)
	return n.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Abort(ErrorRequired(ctx))
		}
	})
}

// Optional same as AnySchema.Optional
func (n *NumberSchema) Optional() *NumberSchema {
	n.required = boolPtr(false)
	return n.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Skip()
		}
	})
}

// Default same as AnySchema.Default
func (n *NumberSchema) Default(value float64) *NumberSchema {
	n.required = boolPtr(false)
	return n.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Value = value
		}
	})
}

// Set same as AnySchema.Set
func (n *NumberSchema) Set(value float64) *NumberSchema {
	return n.Transform(func(ctx *Context) {
		ctx.Value = value
	})
}

// Equal same as AnySchema.Equal
func (n *NumberSchema) Equal(value float64) *NumberSchema {
	return n.Check(func(ctxValue float64) error {
		if value != ctxValue {
			return errors.New(ErrorMessageEqual(value))
		}
		return nil
	})
}

// When same as AnySchema.When
func (n *NumberSchema) When(refPath string, condition interface{}, then Schema) *NumberSchema {
	return n.Transform(func(ctx *Context) { n.whenEqual(ctx, refPath, condition, then) })
}

// Check use the provided function to validate the value of the key.
// Throws an error whenEqual the value is not float64.
func (n *NumberSchema) Check(f func(float64) error) *NumberSchema {
	return n.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(float64)
		if !ok {
			ctx.Abort(ErrorTypeNumber(ctx))
			return
		}
		if err := f(ctxValue); err != nil {
			ctx.ErrorBag.Add(NewError(ctx, err.Error()))
		}
	})
}

// Valid same as AnySchema.Valid
func (n *NumberSchema) Valid(values ...float64) *NumberSchema {
	return n.Check(func(ctxValue float64) error {
		var isValid bool
		for _, v := range values {
			if v == ctxValue {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New(ErrorMessageNumberOneOf(values))
		}
		return nil
	})
}

// Min check if the value is greater than or equal to the provided value.
func (n *NumberSchema) Min(min float64) *NumberSchema {
	return n.Check(func(ctxValue float64) error {
		if ctxValue < min {
			return errors.New(ErrorMessageMin(min))
		}
		return nil
	})
}

// Max check if the value is less than or equal to the provided value.
func (n *NumberSchema) Max(max float64) *NumberSchema {
	return n.Check(func(ctxValue float64) error {
		if ctxValue > max {
			return errors.New(ErrorMessageMax(max))
		}
		return nil
	})
}

// GreaterThanOrEqualToField checks if the value is greater than or equal to the value at `refPath`
func (n *NumberSchema) GreaterThanOrEqualToField(refPath string) *NumberSchema {
    return n.Transform(func (ctx *Context) {
        ctxValue, ok := ctx.Value.(float64)
        if !ok {
            ctx.Abort(ErrorTypeNumber(ctx))
            return
        }

        r, _ := ctx.Ref(refPath)
        refValue, ok := r.(float64)
        if !ok {
            ctx.ErrorBag.AddFromContext(ctx, ErrorMessageTypeNumber())
            return
        }

        if ctxValue < refValue {
            ctx.ErrorBag.Add(ErrorMin(ctx, refPath))
        }
    })
}

// Integer check if the value is integer.
func (n *NumberSchema) Integer() *NumberSchema {
	return n.Check(func(ctxValue float64) error {
		if ctxValue != math.Trunc(ctxValue) {
			return errors.New(ErrorMessageTypeInt())
		}
		return nil
	})
}

// Convert use the provided function to convert the value of the key.
// Throws an error whenEqual the value is not float64.
func (n *NumberSchema) Convert(f func(float64) float64) *NumberSchema {
	return n.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(float64)
		if !ok {
			ctx.Abort(ErrorTypeNumber(ctx))
			return
		}
		ctx.Value = f(ctxValue)
	})
}

// Ceil convert the value to the least integer value greater than or equal to the value.
func (n *NumberSchema) Ceil() *NumberSchema {
	return n.Convert(math.Ceil)
}

// Floor convert the value to the greatest integer value less than or equal to the value.
func (n *NumberSchema) Floor() *NumberSchema {
	return n.Convert(math.Floor)
}

// Round convert the value to the nearest integer, rounding half away from zero.
func (n *NumberSchema) Round() *NumberSchema {
	return n.Convert(math.Round)
}

// ParseString convert the string value to float64.
// Validation will be skipped whenEqual this value is not string.
// But if this value is not a valid number, an error will be thrown.
func (n *NumberSchema) ParseString() *NumberSchema {
	return n.Transform(func(ctx *Context) {
		if ctxValue, ok := ctx.Value.(string); ok {
			value, err := strconv.ParseFloat(ctxValue, 64)
			if err != nil {
				ctx.Abort(ErrorTypeNumber(ctx))
				return
			}
			ctx.Value = value
		}
	})
}

// Validate same as AnySchema.Validate
func (n *NumberSchema) Validate(ctx *Context) {
	if n.required == nil {
		n.Optional()
	}
	if ctxValue, ok := ctx.Value.(int); ok {
		ctx.Value = float64(ctxValue)
	}
	for _, rule := range n.rules {
		rule(ctx)
		if ctx.skip {
			return
		}
	}
	if ctx.ErrorBag.Empty() {
		if _, ok := (ctx.Value).(float64); !ok {
			ctx.Abort(ErrorTypeNumber(ctx))
		}
	}
}
