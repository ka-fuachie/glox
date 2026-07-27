package interpreter

import "fmt"

type Return struct {
	Value any
}

func NewReturn(value any) *Return {
	return &Return{Value: value}
}

func (r *Return) Error() string {
	return fmt.Sprintf("%v", *r)
}
