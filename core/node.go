package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// "fmt"

type ResultType byte

const (
	Boolean ResultType = iota
	String
	Int
	Failure
	Unknown
)

func (n ResultType) String() string {
	switch n {
	case Boolean:
		return "Boolean"
	case String:
		return "String"
	case Int:
		return "Int"
	case Failure:
		return "Failure"
	case Unknown:
		return "Unknown"
	}
	panic(fmt.Sprintf("unknown ResultType: %d", n))
}

type Kind byte

const (
	Expression = iota
	Atom
	ConditionAtom
	Declaration
	SymbolAtom
	BinaryOperator
	ArgumentVariable
)

func (n Kind) String() string {
	switch n {
	case Expression:
		return "Expression"
	case Atom:
		return "Atom"
	case ConditionAtom:
		return "ConditionAtom"
	case Declaration:
		return "Declaration"
	case SymbolAtom:
		return "SymbolAtom"
	case BinaryOperator:
		return "BinaryOperator"
	case ArgumentVariable:
		return "ArgumentVariable"
	}
	panic(fmt.Sprintf("unknown Kind: %d", n))
}

type Node struct {
	Nodes         []*Node
	Kind          Kind
	Value         Value
	ArgumentIndex int
	Location      *Location
	Type          ResultType
}

func node(kind Kind, token *Token, rt ResultType, nodes []*Node) *Node {
	var node Node
	node.Type = rt
	node.Kind = kind
	node.Value = token.Value
	node.Nodes = nodes
	node.Location = &token.Location
	//log.Println("node", fmt.Sprintf("%p", &node), ": kind ", kind, " | type ", rt, "| value ", value_to_any(token.Value), " | nodes", nodes)
	return &node
}

func (n *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":  n.Type,
		"kind":  n.Kind,
		"value": value_to_any(n.Value),
		"nodes": n.Nodes,
	})
}

func (n *Node) String() string {
	sb := strings.Builder{}
	_print(&sb, n, "")
	return sb.String()
}

func _print(sb *strings.Builder, node *Node, indent string) {
	sb.WriteString(fmt.Sprintln(indent, node.Kind, "-", node.Type, ":", node.Value, "[", value_to_any(node.Value), "]"))
	for _, n := range node.Nodes {
		_print(sb, n, indent+"  ")
	}
}
