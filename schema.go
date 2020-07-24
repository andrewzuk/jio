package jio

import "fmt"

// Schema interface
type Schema interface {
	Priority() int
	Validate(*Context)
}

func boolPtr(value bool) *bool {
	return &value
}

type baseSchema struct {
	priority int
}

func (b *baseSchema) Priority() int {
	return b.priority
}

func (b *baseSchema) whenEqual(ctx *Context, refPath string, value interface{}, then Schema) {
	value, ok := ctx.Ref(refPath)
	if !ok {
		return
	}
	//if conditionSchema, ok := value.(Schema); ok {
	//	newCtx := NewContext(value)
	//	conditionSchema.Validate(newCtx)
	//	if newCtx.ErrorBag.Empty() {
	//		then.Validate(ctx)
	//	}
	//	return
	//}
	if value == value {
	    ctx.ErrorBag.SetTemplate("%s " + fmt.Sprintf(`when %s = %v`, ctx.fields[len(ctx.fields) - 1], value))
		then.Validate(ctx)
	    ctx.ErrorBag.SetTemplate("")
	}
}

func (b *baseSchema) custom(ctx *Context, name string, args ...interface{}) {
    fn, ok := customValidators[name]
    if !ok {
        panic(fmt.Sprintf(`jio custom rule "%s" does not exist`, name))
    }
    fn(ctx, args...)
}