package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// "fmt"

type ResultType string

const (
	Boolean ResultType = "Boolean"
	String  ResultType = "String"
	Int     ResultType = "Int"
	Failure ResultType = "Failure"
	Unknown ResultType = "Unknown"
)

type Kind string

const (
	Expression       Kind = "Expression"
	Atom             Kind = "Atom"
	ConditionAtom    Kind = "ConditionAtom"
	Declaration      Kind = "Declaration"
	SymbolAtom       Kind = "SymbolAtom"
	BinaryOperator   Kind = "BinaryOperator"
	ArgumentVariable Kind = "ArgumentVariable"
)

type Node struct {
	Kind          Kind
	Type          ResultType
	Value         Value
	Nodes         []*Node
	ArgumentIndex int
	Location      *Location
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
