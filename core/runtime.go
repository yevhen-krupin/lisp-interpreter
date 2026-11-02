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

type EvaluationResult struct {
	Type  ResultType
	Value Value
	Error error
}

type Argument struct {
	Value any
	Type  ResultType
}

type StackFrame struct {
	id        uint64
	Arguments []Value
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
	switch node.Kind {
	//literals
	case Atom:
		return EvaluationResult{node.Type, node.Value, nil}
	case ConditionAtom:
		if len(node.Nodes) != 3 {
			return EvaluationResult{Type: Unknown, Error: fmt.Errorf("condition atom should have 3 children nodes: condition body and two branches for true and false")}
		}
		result := r.evaluate(node.Nodes[0], frame)
		if result.Error != nil {
			return result
		}
		if result.Type != Boolean {
			return EvaluationResult{Type: Unknown, Error: fmt.Errorf("condition evaluation should to return boolean but was %v : %v", result.Type, result.Value)}
		}
		if result.Value.ValueBoolean {
			//log.Println("condition -> true branch")
			return r.evaluate(node.Nodes[1], frame)
		} else {
			//log.Println("condition -> false branch", node.Nodes[2])
			return r.evaluate(node.Nodes[2], frame)
		}
	case ArgumentVariable:
		if len(node.Nodes) == 0 && node.ArgumentIndex >= 0 && node.ArgumentIndex < len(frame.Arguments) {
			// todo: convert into a real stack
			arg_value := frame.Arguments[node.ArgumentIndex]
			res := EvaluationResult{node.Type, arg_value, nil}
			return res
		}
	case Declaration:
		return EvaluationResult{Unknown, ValueString(strings.ToUpper(node.Value.ValueString)), nil}
	case BinaryOperator:
		return r.binary_operator(node, frame)
	case Expression:
		// expression that just returns something
		if len(node.Nodes) == 1 {
			return r.evaluate(node.Nodes[0], frame)
		}
		// empty expression '()' evaluates as nil or false
		if len(node.Nodes) == 0 {
			return EvaluationResult{Boolean, ValueBoolean(false), nil}
		}
		// call function
		if node.Kind == Expression && len(node.Nodes) > 0 && node.Nodes[0].Kind == SymbolAtom {
			if f, ok := (*r.SymbolTable.Functions)[node.Nodes[0].Value.ValueString]; ok {
				return r.call_function(node, f, frame)
			}
		}
	}
	return EvaluationResult{Type: Unknown, Error: fmt.Errorf("unrecognized node")}
}

func (r Runtime) binary_operator(oper *Node, frame StackFrame) EvaluationResult {
	//log.Println("evaluating binary operator", oper)
	a := r.operand(oper, 0, frame)
	b := r.operand(oper, 1, frame)

	if a.Error != nil || b.Error != nil {
		return EvaluationResult{Type: Failure, Error: first_error(a.Error, b.Error)}
	}

	if oper.Value.ValueString[0] == '+' {
		return EvaluationResult{Int, ValueInt(a.Value.ValueInt + b.Value.ValueInt), nil}
	} else if oper.Value.ValueString[0] == '-' {
		return EvaluationResult{Int, ValueInt(a.Value.ValueInt - b.Value.ValueInt), nil}
	} else if oper.Value.ValueString[0] == '*' {
		return EvaluationResult{Int, ValueInt(a.Value.ValueInt * b.Value.ValueInt), nil}
	} else if oper.Value.ValueString[0] == '/' {
		return EvaluationResult{Int, ValueInt(a.Value.ValueInt / b.Value.ValueInt), nil}
	} else if oper.Value.ValueString[0] == '=' {
		return EvaluationResult{Boolean, ValueBoolean(a.Value.Equal(b.Value)), nil}
	} else if oper.Value.ValueString[0] == '<' {
		return EvaluationResult{Boolean, ValueBoolean(a.Value.ValueInt < b.Value.ValueInt), nil}
	} else if oper.Value.ValueString[0] == '>' {
		return EvaluationResult{Boolean, ValueBoolean(a.Value.ValueInt > b.Value.ValueInt), nil}
	}
	return EvaluationResult{Type: Failure, Error: UnknownBinaryOperator(oper.Value.ValueString[0])}
}

func (r Runtime) operand(node *Node, index int, frame StackFrame) EvaluationResult {
	//log.Println("evaluating operand", node.Nodes[index])
	return r.evaluate(node.Nodes[index], frame)
}

// node: symbol atom of call site
// first child is a function, followed by the arguments
func (r Runtime) call_function(callsite *Node, function FunctionEntry, frame StackFrame) EvaluationResult {
	//interrupt to max stack
	resolved := make([]Value, len(callsite.Nodes)-1)
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
		return EvaluationResult{Error: ArgumentsCountError(callsite.Value.ValueString, len(function.Arguments.Nodes), len(resolved))}
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
