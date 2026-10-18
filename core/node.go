package core

import (
	"encoding/json"
	"fmt"
	"log"
)

// "fmt"

type ResultType string

const (
	Boolean ResultType = "Boolean"
	String  ResultType = "String"
	Int     ResultType = "Int"
	Error   ResultType = "Error"
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
	Value         []byte
	Nodes         []*Node
	ArgumentIndex int
}

func node(kind Kind, value []byte, rt ResultType, nodes []*Node) *Node {
	var node Node
	node.Type = rt
	node.Kind = kind
	node.Value = value
	node.Nodes = nodes

	log.Println("node", fmt.Sprintf("%p", &node), ": kind ", kind, " | type ", rt, "| value ", string(value), " | nodes", nodes)
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

func print(node *Node, indent string) {
	log.Println(indent, node.Kind, "-", node.Type, ":", node.Value, "[", value_to_any(node.Value), "]")
	for _, n := range node.Nodes {
		print(n, indent+"  ")
	}
}
