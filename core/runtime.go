package core

import (
	"fmt"
	"strings"
)

type Runtime struct {
	Frame       byte
	SymbolTable *SymbolTable
}

type StackFrame struct {
	Arguments []Value
	id        byte
}

type FunctionEntry struct {
	Body               *Node
	Arguments          *Node
	EvaluatedArguments []Value
}

func NewRuntime(maxStack int) Runtime {
	return Runtime{0, &SymbolTable{Functions: &map[string]*FunctionEntry{}}}
}

func (r *Runtime) EvaluateAst(ast *Node) Value {
	return r.evaluate(ast, &StackFrame{})
}

func (r *Runtime) evaluate(node *Node, frame *StackFrame) Value {
	if node.Kind == Atom {
		return node.Value
	}
	switch node.Kind {
	case ConditionAtom:
		if len(node.Nodes) != 3 {
			return ValueError("condition atom should have 3 children nodes: condition body and two branches for true and false")
		}
		result := r.evaluate(node.Nodes[0], frame)
		if result.T == Failure {
			return result
		}
		if result.T != Boolean {
			return ValueError(fmt.Sprintf("condition evaluation should to return boolean but was %v : %v", result.T, value_to_any(result)))
		}
		if result.ValueInt == 1 {
			//log.Println("condition -> true branch")
			return r.evaluate(node.Nodes[1], frame)
		} else {
			//log.Println("condition -> false branch", node.Nodes[2])
			return r.evaluate(node.Nodes[2], frame)
		}
	case ArgumentVariable:
		if len(node.Nodes) == 0 && node.ArgumentIndex >= 0 && node.ArgumentIndex < len(frame.Arguments) {
			// todo: convert into a real stack
			return frame.Arguments[node.ArgumentIndex]
		}
	case Declaration:
		return ValueString(strings.ToUpper(node.Value.ValueString))
	case BinaryOperator:
		return r.binary_operator(node, frame)
	case CallSite:
		return r.call_function(node, node.Function, frame)
	case Expression:
		// expression that just returns something
		if len(node.Nodes) == 1 {
			return r.evaluate(node.Nodes[0], frame)
		}
		// empty expression '()' evaluates as nil or false
		if len(node.Nodes) == 0 {
			return ValueBoolean(false)
		}
	}
	return node.Value
}

func (r *Runtime) binary_operator(oper *Node, frame *StackFrame) Value {
	//log.Println("evaluating binary operator", oper)
	a := r.operand(oper, 0, frame)
	b := r.operand(oper, 1, frame)
	if a.T == Failure || b.T == Failure {
		return first_error(a, b)
	}
	switch oper.Value.ValueString[0] {
	case '+':
		return ValueInt(a.ValueInt + b.ValueInt)
	case '-':
		return ValueInt(a.ValueInt - b.ValueInt)
	case '*':
		return ValueInt(a.ValueInt * b.ValueInt)
	case '/':
		return ValueInt(a.ValueInt / b.ValueInt)
	case '=':
		return ValueBoolean(a.Equal(b))
	case '<':
		return ValueBoolean(a.ValueInt < b.ValueInt)
	case '>':
		return ValueBoolean(a.ValueInt > b.ValueInt)
	}
	return UnknownBinaryOperator(oper.Value.ValueString[0])
}

func (r *Runtime) operand(node *Node, index int, frame *StackFrame) Value {
	//log.Println("evaluating operand", node.Nodes[index])
	return r.evaluate(node.Nodes[index], frame)
}

// node: symbol atom of call site
// first child is a function, followed by the arguments
func (r *Runtime) call_function(callsite *Node, function *FunctionEntry, frame *StackFrame) Value {
	//interrupt to max stack
	resolved := make([]Value, len(callsite.Nodes))
	for i, n := range callsite.Nodes {
		arg := r.evaluate(n, frame)
		if arg.T != function.Arguments.Nodes[i].Type {
			return TypeConsolidationError(arg, function.Arguments.Nodes[i].Type)
		}
		//log.Println("resolved argument", n, arg)
		resolved[i] = arg
	}
	//log.Println("resolved", len(resolved), resolved)
	frame2 := StackFrame{resolved, frame.id + 1}
	res := r.evaluate(function.Body, &frame2)
	//log.Println("function", callsite.Value.ValueString, callsite.Value.ValueInt, "args", resolved, "result", res.T, ":", res.ValueInt)
	return res
}

func first_error(input ...Value) Value {
	for _, item := range input {
		if item.T == Failure {
			return item
		}
	}
	return ValueError("Unable to find the first error, there are no errors")
}
