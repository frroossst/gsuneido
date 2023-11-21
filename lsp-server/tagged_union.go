package main

import "fmt"

// Option type
type Option struct {
    value interface{}
    valid bool
}

func Some(value interface{}) Option {
    return Option{value, true}
}

var None = Option{nil, false}

func (o Option) Unwrap() interface{} {
    if o.valid {
        return o.value
    }
    panic("called `Option::unwrap()` on a `None` value")
}

func (o Option) IsNone() bool {
	return !o.valid
}

func (o Option) IsSome() bool {
	return o.valid
}

// Result type
type Result struct {
    value interface{}
    err   error
}

func Ok(value interface{}) Result {
    return Result{value, nil}
}

func Err(err error) Result {
    return Result{nil, err}
}

func (r Result) IsOk() bool {
	return r.err == nil
}

func (r Result) IsErr() bool {
	return r.err != nil
}

func (r Result) Unwrap() interface{} {
    if r.err == nil {
        return r.value
    }
    panic(fmt.Sprintf("called `Result::unwrap()` on an `Err` value: %s", r.err))
}
