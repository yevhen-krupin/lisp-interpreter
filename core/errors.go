package core

import (
	"fmt"
	"strings"
)

type Error struct {
	Error    Value
	Location Location
}

type ParserErrors string

func TypeConsolidationError(val Value, expected ResultType) Value {
	return ValueError(fmt.Sprintf("The value %v is not of the expected type %v", value_to_any(val), expected))
}

func UnknownBinaryOperator(operator byte) Value {
	return ValueError(fmt.Sprintf("Unexpected binary operator encountered %v", operator))
}

func ArgumentsCountError(function string, expected int, actual int) Value {
	if actual < expected {
		return ValueError(fmt.Sprintf("Too few arguments in call to %v\n%v arguments provided, at least %v required", string(function), actual, expected))
	} else {
		return ValueError(fmt.Sprintf("Too many arguments in call to %v\n%v arguments provided, at most %v required", string(function), actual, expected))
	}
}

func UnexpectedStackStateError() Value {
	return ValueError("unexpected state: the stack frame after leaving the function is not correct")
}

func StackOverflow(depth byte) Value {
	return ValueError(fmt.Sprintf("Stack overflow at size %v. Stack can probably be resized.", depth))
}

type ErrorStream []Error

func (p ErrorStream) Diagnostic() string {
	var sb strings.Builder
	for _, e := range p {
		sb.WriteString(fmt.Sprintln(":", e.Error, e.Location))
	}
	return sb.String()
}
