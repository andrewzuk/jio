package jio

import (
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "strings"
)

type FieldError struct {
    Field string
    Err   error
}

func (err FieldError) Error() string {
    return fmt.Sprintf(`%s %s`, err.Field, err.Err.Error())
}

type ErrorBag struct {
    m map[FieldError]interface{}
    tmpl string
}

// TODO: nested array paths not correct

func NewErrorBag() *ErrorBag {
    return &ErrorBag{m: map[FieldError]interface{}{}}
}

func (bag *ErrorBag) SetTemplate(template string) {
    bag.tmpl = template
}

func applyErrorMessageTemplate(template string, err FieldError) FieldError {
    if template == "" {
        return err
    }

    err.Err = fmt.Errorf(template, err.Err.Error())

    return err
}

func (bag *ErrorBag) Add(err FieldError) {
    if bag.tmpl != "" {
        err = applyErrorMessageTemplate(bag.tmpl, err)
    }
    bag.m[err] = nil
}

func (bag *ErrorBag) AddFromContext(ctx *Context, str string) {
    bag.Add(FieldError{ctx.FieldPath(), errors.New(str)})
}

func (bag *ErrorBag) AddBag(bag2 *ErrorBag) {
    for err := range bag2.m {
        bag.Add(err)
    }
}

func (bag *ErrorBag) Empty() bool {
    return len(bag.m) == 0
}

func (bag *ErrorBag) StringArray() []string {
    var list []string
    for err := range bag.m {
        list = append(list, err.Error())
    }
    sort.Strings(list)
    return list
}

func (bag *ErrorBag) MarshalJSON() ([]byte, error) {
    return json.Marshal(bag.StringArray())
}

func (bag *ErrorBag) Error() string {
    return fmt.Sprintf("[%s]", strings.Join(bag.StringArray(), "; "))
}

func NewError(ctx *Context, msg string) FieldError {
    return FieldError{ctx.FieldPath(), errors.New(msg)}
}

func ErrorStringLengthMin(ctx *Context, min int) FieldError {
    return NewError(ctx, ErrorMessageStringLengthMin(min))
}

func ErrorMessageStringLengthMin(min int) string {
    return fmt.Sprintf(`must have at least %s`, characters(min))
}

func ErrorStringLengthMax(ctx *Context, max int) FieldError {
    return NewError(ctx, ErrorMessageStringLengthMax(max))
}

func ErrorMessageStringLengthMax(max int) string {
    return fmt.Sprintf(`cannot have more than %s`, characters(max))
}

func ErrorStringLengthEqual(ctx *Context, val int) FieldError {
    return NewError(ctx, ErrorMessageStringLengthEqual(val))
}

func ErrorMessageStringLengthEqual(val int) string {
    return fmt.Sprintf(`must have exactly %s`, characters(val))
}

func ErrorArrayLengthMin(ctx *Context, min int) FieldError {
    return NewError(ctx, ErrorMessageArrayLengthMin(min))
}

func ErrorMessageArrayLengthMin(min int) string {
    return fmt.Sprintf(`must have at least %s`, items(min))
}

func ErrorArrayLengthMax(ctx *Context, max int) FieldError {
    return NewError(ctx, ErrorMessageArrayLengthMax(max))
}

func ErrorMessageArrayLengthMax(max int) string {
    return fmt.Sprintf(`cannot have more than %s`, items(max))
}

func ErrorArrayLengthEqual(ctx *Context, val int) FieldError {
    return NewError(ctx, ErrorMessageArrayLengthEqual(val))
}

func ErrorMessageArrayLengthEqual(val int) string {
    return fmt.Sprintf(`must have exactly %s`, items(val))
}

func ErrorMessageArrayUniqueObjects(fields []string) string {
    return fmt.Sprintf("must be unique [fields: %s]", strings.Join(fields, ", "))
}

func ErrorArrayUniqueObjects(ctx *Context, fields []string) FieldError {
    return NewError(ctx, ErrorMessageArrayUniqueObjects(fields))
}

func ErrorMin(ctx *Context, min interface{}) FieldError {
    return NewError(ctx, ErrorMessageMin(min))
}

func ErrorMessageMin(min interface{}) string {
    return fmt.Sprintf(`must be >= %v`, min)
}

func ErrorMax(ctx *Context, min interface{}) FieldError {
    return NewError(ctx, ErrorMessageMax(min))
}

func ErrorMessageMax(max interface{}) string {
    return fmt.Sprintf(`must be <= %v`, max)
}

func ErrorEqual(ctx *Context, val interface{}) FieldError {
    return NewError(ctx, ErrorMessageEqual(val))
}

func ErrorMessageEqual(val interface{}) string {
    return fmt.Sprintf(`must equal %v`, val)
}

func ErrorRequired(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageRequired())
}

func ErrorMessageRequired() string {
    return fmt.Sprintf(`is required`)
}

func ErrorType(ctx *Context, t string) FieldError {
    return NewError(ctx, ErrorMessageType(t))
}

func ErrorMessageType(t string) string {
    return fmt.Sprintf(`must be %s`, t)
}

func ErrorTypeObject(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageTypeObject())
}

func ErrorMessageTypeObject() string {
    return ErrorMessageType("an object")
}

func ErrorTypeArray(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageTypeArray())
}

func ErrorMessageTypeArray() string {
    return ErrorMessageType("an array")
}

func ErrorTypeString(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageTypeString())
}

func ErrorMessageTypeString() string {
    return ErrorMessageType("a string")
}

func ErrorTypeBool(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageTypeBool())
}

func ErrorMessageTypeBool() string {
    return ErrorMessageType("a boolean")
}

func ErrorTypeInt(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageTypeInt())
}

func ErrorMessageTypeInt() string {
    return ErrorMessageType("an integer")
}

func ErrorTypeNumber(ctx *Context) FieldError {
    return NewError(ctx, ErrorMessageTypeNumber())
}

func ErrorMessageTypeNumber() string {
    return ErrorMessageType("a number")
}

func ErrorOneOf(ctx *Context, values []interface{}) FieldError {
    return NewError(ctx, ErrorMessageOneOf(values))
}

func ErrorMessageOneOf(values []interface{}) string {
    var vals []string
    for _, v := range values {
        vals = append(vals, fmt.Sprintf(`%v`, v))
    }
    return fmt.Sprintf(`must be one of [%s]`, strings.Join(vals, ", "))
}

func ErrorNotOneOf(ctx *Context, values []interface{}) FieldError {
    return NewError(ctx, ErrorMessageNotOneOf(values))
}

func ErrorMessageNotOneOf(values []interface{}) string {
    var vals []string
    for _, v := range values {
        vals = append(vals, fmt.Sprintf(`%s`, v))
    }
    return fmt.Sprintf(`cannot be any of [%s]`, strings.Join(vals, ", "))
}

func ErrorMessageNumberOneOf(values []float64) string {
    var list []interface{}
    for _, v := range values {
        list = append(list, interface{}(v))
    }
    return ErrorMessageOneOf(list)
}

func ErrorStringOneOf(ctx *Context, values []string) FieldError {
    return NewError(ctx, ErrorMessageStringOneOf(values))
}

func ErrorMessageStringOneOf(values []string) string {
    var list []interface{}
    for _, v := range values {
        list = append(list, interface{}(v))
    }
    return ErrorMessageOneOf(list)
}

func ErrorMatchPattern(ctx *Context, pattern string) FieldError {
    return NewError(ctx, ErrorMessageMatchPattern(pattern))
}

func ErrorMessageMatchPattern(pattern string) string {
    return fmt.Sprintf(`must match pattern %s`, pattern)
}

func ErrorObjectMissingRequiredKeys(ctx *Context, missingKeys []string) FieldError {
    return NewError(ctx,  ErrorMessageObjectMissingRequiredKeys(missingKeys))
}

func ErrorMessageObjectMissingRequiredKeys(missingKeys []string) string {
    return fmt.Sprintf(`is missing required keys [%s]`, strings.Join(missingKeys, ", "))
}

func ErrorObjectContainsForbiddenKeys(ctx *Context, forbiddenKeys []string) FieldError {
    return NewError(ctx,  ErrorMessageObjectContainsForbiddenKeys(forbiddenKeys))
}

func ErrorMessageObjectContainsForbiddenKeys(missingKeys []string) string {
    return fmt.Sprintf(`contains forbidden keys [%s]`, strings.Join(missingKeys, ", "))
}

func characters(count int) string {
    str := "character"
    if count != 1 {
        str += "s"
    }
    return fmt.Sprintf(`%d %s`, count, str)
}

func items(count int) string {
    str := "item"
    if count != 1 {
        str += "s"
    }
    return fmt.Sprintf(`%d %s`, count, str)
}