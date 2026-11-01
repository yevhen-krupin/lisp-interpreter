package core

import (
	"fmt"
	"strings"
	"sync/atomic"
)

var ops atomic.Uint64

type Runtime struct {
	Stack       []StackFrame
	Frame       int
	SymbolTable *SymbolTable
}

type EvaluationResultValue any

type EvaluationResult struct {
	Type  ResultType
	Value EvaluationResultValue
	Error error
}

type Argument struct {
	Value any
	Type  ResultType
}

type StackFrame struct {
	id        uint64
	Arguments []EvaluationResultValue
}

type FunctionEntry struct {
	Name               string
	Body               *Node
	Arguments          *Node
	EvaluatedArguments []EvaluationResult
}

func NewRuntime(maxStack int) Runtime {
	return Runtime{make([]StackFrame, maxStack), 0, &SymbolTable{Functions: &map[string]FunctionEntry{}}}
}

func (r Runtime) EvaluateAst(ast *Node) EvaluationResult {
	return r.evaluate(ast, StackFrame{})
}

func (r Runtime) evaluate(node *Node, frame StackFrame) EvaluationResult {
	// int literal
	if node.Kind == Atom && node.Type == Int {
		return EvaluationResult{Int, bytes_to_int32(node.Value), nil}
	}
	// string literal
	if node.Kind == Atom && node.Type == String {
		return EvaluationResult{String, string(node.Value), nil}
	}
	// boolean literal
	if node.Kind == Atom && node.Type == Boolean {
		return EvaluationResult{Boolean, node.Value[0] == 1, nil}
	}
	// condition
	if node.Kind == ConditionAtom {
		if len(node.Nodes) != 3 {
			return EvaluationResult{Unknown, nil, fmt.Errorf("condition atom should have 3 children nodes: condition body and two branches for true and false")}
		}
		result := r.evaluate(node.Nodes[0], frame)
		if result.Error != nil {
			return result
		}
		if result.Type != Boolean {
			return EvaluationResult{Unknown, nil, fmt.Errorf("condition evaluation should to return boolean but was %v : %v", result.Type, result.Value)}
		}
		if result.Value == true {
			//log.Println("condition -> true branch")
			return r.evaluate(node.Nodes[1], frame)
		} else {
			//log.Println("condition -> false branch", node.Nodes[2])
			return r.evaluate(node.Nodes[2], frame)
		}
	}

	// resolve argument variable
	if node.Kind == ArgumentVariable && len(node.Nodes) == 0 && node.ArgumentIndex >= 0 && node.ArgumentIndex < len(frame.Arguments) {
		// todo: convert into a real stack
		arg_value := frame.Arguments[node.ArgumentIndex]
		return EvaluationResult{node.Type, arg_value, nil}
	}
	// call function
	if node.Kind == Expression && len(node.Nodes) > 0 && node.Nodes[0].Kind == SymbolAtom {
		if f, ok := (*r.SymbolTable.Functions)[string(node.Nodes[0].Value)]; ok {
			return r.call_function(node, f, frame)
		}
	}

	// operators
	if node.Kind == BinaryOperator {
		oper := node
		return r.binary_operator(oper, frame)
	}

	// expression that just returns something
	if node.Kind == Expression && len(node.Nodes) == 1 {
		return r.evaluate(node.Nodes[0], frame)
	}

	// empty expression evaluates as nil or false
	if node.Kind == Expression && len(node.Nodes) == 0 {
		return EvaluationResult{Boolean, false, nil}
	}

	if node.Kind == Declaration {
		return EvaluationResult{Unknown, strings.ToUpper(string(node.Value)), nil}
	}

	return EvaluationResult{Unknown, nil, fmt.Errorf("unrecognized node %v", node)}
}

func (r Runtime) binary_operator(oper *Node, frame StackFrame) EvaluationResult {
	//log.Println("evaluating binary operator", oper)
	a := r.operand(oper, 0, frame)
	b := r.operand(oper, 1, frame)

	if a.Error != nil || b.Error != nil {
		return EvaluationResult{Failure, nil, first_error(a.Error, b.Error)}
	}

	if oper.Value[0] == '+' {
		return EvaluationResult{Int, a.Value.(int) + b.Value.(int), nil}
	}
	if oper.Value[0] == '-' {
		return EvaluationResult{Int, a.Value.(int) - b.Value.(int), nil}
	}
	if oper.Value[0] == '*' {
		return EvaluationResult{Int, a.Value.(int) * b.Value.(int), nil}
	}
	if oper.Value[0] == '/' {
		return EvaluationResult{Int, a.Value.(int) / b.Value.(int), nil}
	}
	if oper.Value[0] == '=' {
		return EvaluationResult{Boolean, a.Value.(int) == b.Value.(int), nil}
	}
	if oper.Value[0] == '<' {
		return EvaluationResult{Boolean, a.Value.(int) < b.Value.(int), nil}
	}
	if oper.Value[0] == '>' {
		return EvaluationResult{Boolean, a.Value.(int) > b.Value.(int), nil}
	}
	return EvaluationResult{Failure, nil, UnknownBinaryOperator(oper.Value[0])}
}

func (r Runtime) operand(node *Node, index int, frame StackFrame) EvaluationResult {
	//log.Println("evaluating operand", node.Nodes[index])
	return r.evaluate(node.Nodes[index], frame)
}

// node: symbol atom of call site
// first child is a function, followed by the arguments
func (r Runtime) call_function(callsite *Node, function FunctionEntry, frame StackFrame) EvaluationResult {
	//interrupt to max stack
	resolved := make([]EvaluationResultValue, len(callsite.Nodes)-1)
	for i, n := range callsite.Nodes {
		if i > 0 {
			arg := r.evaluate(n, frame)
			if arg.Type != function.Arguments.Nodes[i-1].Type {
				return EvaluationResult{Error: TypeConsolidationError(arg.Value, function.Arguments.Nodes[i-1].Type)}
			}
			//log.Println("resolved argument", n, arg)
			resolved[i-1] = arg.Value
		}
	}
	if r.Frame+1 == len(r.Stack) {
		return EvaluationResult{Error: StackOverflow(r.Frame + 1)}
	}
	// check arguments
	if len(function.Arguments.Nodes) != len(resolved) {
		return EvaluationResult{Error: ArgumentsCountError(callsite.Value, len(function.Arguments.Nodes), len(resolved))}
	}

	//log.Println("calling function", name, "args", args, "body", functionBody)
	frame2 := StackFrame{ops.Add(1), resolved}
	res := r.evaluate(function.Body, frame2)
	//log.Println("function", name, "args", args, "result", res.Type, ":", res.Value)
	return res
}

func first_error(input ...error) error {
	for _, item := range input {
		if item != nil {
			return item
		}
	}
	return nil
}

func (e EvaluationResult) String() string {
	return fmt.Sprintln(e.Value)
	//return fmt.Sprintln("Evaluated @ type:", e.Type, "value:", e.Value, "error:", e.Error)
}
