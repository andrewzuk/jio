package jio

import (
	"sort"
)

type objectItem struct {
	key    string
	schema Schema
}

// K object keys schema alias
type K map[string]Schema

func (k K) sort() []objectItem {
	objects := make([]objectItem, 0, len(k))
	for key, schema := range k {
		objects = append(objects, objectItem{key, schema})
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].schema.Priority() > objects[j].schema.Priority()
	})
	return objects
}

// Object Generates a schema object that matches object data type
func Object() *ObjectSchema {
	return &ObjectSchema{
		rules: make([]func(*Context), 0, 3),
	}
}

var _ Schema = new(ObjectSchema)

// ObjectSchema match object data type
type ObjectSchema struct {
	baseSchema

	required *bool
	rules    []func(*Context)
}

// SetPriority same as AnySchema.SetPriority
func (o *ObjectSchema) SetPriority(priority int) *ObjectSchema {
	o.priority = priority
	return o
}

// PrependTransform same as AnySchema.PrependTransform
func (o *ObjectSchema) PrependTransform(f func(*Context)) *ObjectSchema {
	o.rules = append([]func(*Context){f}, o.rules...)
	return o
}

// Transform same as AnySchema.Transform
func (o *ObjectSchema) Transform(f func(*Context)) *ObjectSchema {
	o.rules = append(o.rules, f)
	return o
}

// Custom adds a custom validation
func (o *ObjectSchema) Custom(name string, args ...interface{}) *ObjectSchema {
    return o.Transform(func(ctx *Context) {
        o.baseSchema.custom(ctx, name, args...)
    })
}

// Required same as AnySchema.Required
func (o *ObjectSchema) Required() *ObjectSchema {
	o.required = boolPtr(true)
	return o.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Abort(ErrorRequired(ctx))
		}
	})
}

// Optional same as AnySchema.Optional
func (o *ObjectSchema) Optional() *ObjectSchema {
	o.required = boolPtr(false)
	return o.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Skip()
		}
	})
}

// Default same as AnySchema.Default
func (o *ObjectSchema) Default(value map[string]interface{}) *ObjectSchema {
	o.required = boolPtr(false)
	return o.PrependTransform(func(ctx *Context) {
		if ctx.Value == nil {
			ctx.Value = value
		}
	})
}

// With require the presence of these keys.
func (o *ObjectSchema) With(keys ...string) *ObjectSchema {
	return o.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(map[string]interface{})
		if !ok {
			ctx.Abort(ErrorTypeObject(ctx))
			return
		}

		var missingKeys []string
		for _, key := range keys {
			_, ok := ctxValue[key]
			if !ok {
			    missingKeys = append(missingKeys, key)
			}
		}
		if len(missingKeys) > 0 {
		    ctx.ErrorBag.Add(ErrorObjectMissingRequiredKeys(ctx, missingKeys))
        }
	})
}

// Without forbids the presence of these keys.
func (o *ObjectSchema) Without(keys ...string) *ObjectSchema {
	return o.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(map[string]interface{})
		if !ok {
			ctx.ErrorBag.Add(ErrorTypeObject(ctx))
		}
		forbiddenKeys := make([]string, 0, 3)
		for _, key := range keys {
			_, ok := ctxValue[key]
			if ok {
				forbiddenKeys = append(forbiddenKeys, key)
			}
		}
		if len(forbiddenKeys) > 1 {
			ctx.ErrorBag.Add(ErrorObjectContainsForbiddenKeys(ctx, forbiddenKeys))
			return
		}
	})
}

// When same as AnySchema.When
func (o *ObjectSchema) When(refPath string, condition interface{}, then Schema) *ObjectSchema {
	return o.Transform(func(ctx *Context) { o.whenEqual(ctx, refPath, condition, then) })
}

// Keys set the object keys's schema
func (o *ObjectSchema) Keys(children K) *ObjectSchema {
	return o.Transform(func(ctx *Context) {
		ctxValue, ok := ctx.Value.(map[string]interface{})
		if !ok {
			ctx.Abort(ErrorTypeObject(ctx))
			return
		}
		fields := make([]string, len(ctx.fields))
		copy(fields, ctx.fields)

		defer func() {
			ctx.fields = fields
			ctx.Value = ctxValue
		}()

		for _, obj := range children.sort() {
			value, _ := ctxValue[obj.key]
			ctx.parent = ctxValue
			ctx.skip = false
			ctx.fields = append(fields, obj.key)
			ctx.Value = value
			if _, ok := obj.schema.(*ObjectSchema); ok && ctx.parentRoot == nil {
			    ctx.parentRoot = ctx.parent
            }
			obj.schema.Validate(ctx)
			if ctx.ErrorBag.Empty() && !ctx.skip {
				ctxValue[obj.key] = ctx.Value
			}
		}
	})
}

// Validate same as AnySchema.Validate
func (o *ObjectSchema) Validate(ctx *Context) {
    if ctx.Value != nil {
        if _, ok := (ctx.Value).(map[string]interface{}); !ok {
            ctx.Abort(ErrorTypeObject(ctx))
            return
        }
    }
	if o.required == nil {
		o.Optional()
	}
	for _, rule := range o.rules {
		rule(ctx)
		if ctx.skip {
			return
		}
	}
}
