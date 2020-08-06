package jio

import (
    "errors"
    "fmt"
    "reflect"
)

var _ Schema = new(ArraySchema)

// Array Generates a schema object that matches array data type
func Array() *ArraySchema {
	return &ArraySchema{
		rules: make([]func(*Context), 0, 3),
	}
}

// ArraySchema match array data type
type ArraySchema struct {
	baseSchema

	required *bool
	rules    []func(*Context)
}

// SetPriority same as AnySchema.SetPriority
func (a *ArraySchema) SetPriority(priority int) *ArraySchema {
	a.priority = priority
	return a
}

// PrependTransform same as AnySchema.PrependTransform
func (a *ArraySchema) PrependTransform(f func(*Context)) *ArraySchema {
	a.rules = append([]func(*Context){f}, a.rules...)
	return a
}

// Transform same as AnySchema.Transform
func (a *ArraySchema) Transform(f func(*Context)) *ArraySchema {
	a.rules = append(a.rules, f)
	return a
}

// Custom adds a custom validation
func (a *ArraySchema) Custom(name string, args ...interface{}) *ArraySchema {
    return a.Transform(func(ctx *Context) {
        a.baseSchema.custom(ctx, name, args...)
    })
}

// Required same as AnySchema.Required
func (a *ArraySchema) Required() *ArraySchema {
	a.required = boolPtr(true)
	return a.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Abort(ErrorRequired(ctx))
		}
	})
}

// Optional same as AnySchema.Optional
func (a *ArraySchema) Optional() *ArraySchema {
	a.required = boolPtr(false)
	return a.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Skip()
		}
	})
}

// Default same as AnySchema.Default
func (a *ArraySchema) Default(value interface{}) *ArraySchema {
	a.required = boolPtr(false)
	return a.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Value = value
		}
	})
}

// When same as AnySchema.When
func (a *ArraySchema) When(refPath string, condition interface{}, then Schema) *ArraySchema {
	return a.Transform(func(ctx *Context) { a.whenEqual(ctx, refPath, condition, then) })
}

// Check use the provided function to validate the value of the key.
// Throws an error whenEqual the value is not a slice.
func (a *ArraySchema) Check(f func(*Context) error) *ArraySchema {
	return a.Transform(func(ctx *Context) {
		if !ctx.AssertKind(reflect.Slice) {
			ctx.Abort(ErrorTypeArray(ctx))
			return
		}
		err := f(ctx)
		if err != nil {
            if bag, ok := err.(*ErrorBag); ok {
                ctx.ErrorBag.AddBag(bag)
            } else if ferr, ok := err.(FieldError); ok {
                ctx.ErrorBag.Add(ferr)
            } else {
                ctx.ErrorBag.Add(FieldError{ctx.FieldPath(), err})
            }
        }
	})
}

// Items check if this value can pass the validation of any schema.
func (a *ArraySchema) Items(schemas ...Schema) *ArraySchema {
	return a.Check(func(ctx *Context) error {
		ctxRV := reflect.ValueOf(ctx.Value)
		errs := NewErrorBag()
		for i := 0; i < ctxRV.Len(); i++ {
			rv := ctxRV.Index(i).Interface()
			for _, schema := range schemas {
				ctxNew := NewContext(rv)
				ctxNew.root = ctx.root
				ctxNew.parentRoot = ctx.parentRoot
				ctxNew.parent = ctx.parent
				ctxNew.fields = append(ctxNew.fields, append(ctx.fields, fmt.Sprintf(`%d`, i))...)
				schema.Validate(ctxNew)
				errs.AddBag(ctxNew.ErrorBag)
			}
		}
		return errs
	})
}

// Min check if the length of this slice is greater than or equal to the provided length.
func (a *ArraySchema) Min(min int) *ArraySchema {
	return a.Check(func(ctx *Context) error {
		if reflect.ValueOf(ctx.Value).Len() < min {
			return errors.New(ErrorMessageArrayLengthMin(min))
		}
		return nil
	})
}

// Max check if the length of this slice is less than or equal to the provided length.
func (a *ArraySchema) Max(max int) *ArraySchema {
	return a.Check(func(ctx *Context) error {
		if reflect.ValueOf(ctx.Value).Len() > max {
			return errors.New(ErrorMessageArrayLengthMax(max))
		}
		return nil
	})
}

// Length check if the length of this slice is equal to the provided length.
func (a *ArraySchema) Length(length int) *ArraySchema {
	return a.Check(func(ctx *Context) error {
		if reflect.ValueOf(ctx.Value).Len() != length {
			return errors.New(ErrorMessageArrayLengthEqual(length))
		}
		return nil
	})
}

// UniqueObjects checks that all slice objects are unique, using a concatenation of all field values as the composite key for each object.
func (a *ArraySchema) UniqueObjects(fields ...string) *ArraySchema {
    return a.Check(func(ctx *Context) error {
        ref := reflect.ValueOf(ctx.Value)
        valsMap := map[interface{}]bool{}
        for i := 0; i < ref.Len(); i++ {
            obj, ok := ref.Index(i).Interface().(map[string]interface{})
            if !ok {
                // only works for a list of objects at present
                return nil
            }

            var key string
            for _, field := range fields {
                key += obj[field].(string)
            }

            if _, ok := valsMap[key]; ok {
                // not unique
                return errors.New(ErrorMessageArrayUniqueObjects(fields))
            }

            valsMap[key] = true
        }

        return nil
    })
}

// Validate same as AnySchema.Validate
func (a *ArraySchema) Validate(ctx *Context) {
    if ctx.Value != nil {
        if !ctx.AssertKind(reflect.Slice) {
            ctx.Abort(ErrorTypeArray(ctx))
            return
        }
    }
	if a.required == nil {
		a.Optional()
	}
	for _, rule := range a.rules {
		rule(ctx)
		if ctx.skip {
			return
		}
	}
}
