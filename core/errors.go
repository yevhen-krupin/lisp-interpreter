package core

import (
	"fmt"
	"strings"
)

type Error struct {
	Error    error
	Location Location
}

type ParserErrors string

func TypeConsolidationError(val any, expected ResultType) error {
	return fmt.Errorf("The value %v is not of the expected type %v", val, expected)
}

func UnknownBinaryOperator(operator byte) error {
	return fmt.Errorf("Unexpected binary operator encountered %v", operator)
}

func ArgumentsCountError(function string, expected int, actual int) error {
	if actual < expected {
		return fmt.Errorf("Too few arguments in call to %v\n%v arguments provided, at least %v required", string(function), actual, expected)
	} else {
		return fmt.Errorf("Too many arguments in call to %v\n%v arguments provided, at most %v required", string(function), actual, expected)
	}
}

func UnexpectedStackStateError() error {
	return fmt.Errorf("unexpected state: the stack frame after leaving the function is not correct")
}

func StackOverflow(depth int) error {
	return fmt.Errorf("Stack overflow at size %v. Stack can probably be resized.", depth)
}

type ErrorStream []Error

func (p ErrorStream) Diagnostic() string {
	var sb strings.Builder
	for _, e := range p {
		sb.WriteString(fmt.Sprintln(":", e.Error, e.Location))
	}
	return sb.String()
}
