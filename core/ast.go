package core

import (
	"fmt"
	"log"
	"slices"
)

type Parser struct {
	SymbolTable *SymbolTable
	Errors      *[]error
	Position    int
	Input       []*Token
	Lexic       *Lexic
}

type SymbolTable struct {
	Functions *map[string]FunctionEntry
}

func ParseAst(input string, symbolTable *SymbolTable, lexic *Lexic) (*Node, TokenStream, []error) {
	data := slices.Collect(tokenize(input, lexic))
	p := Parser{symbolTable, &[]error{}, 0, data, lexic}
	node := p.atom_node()
	return node, TokenStream(data), *p.Errors
}

func (p *Parser) atom_node() *Node {
	return coalesce(
		p,
		// atom can be an expression consisting of other atoms/expressions, it can contain strings and other stuff from below
		(*Parser).expression_node,
		// atom can be a literal: string, int, boolean
		(*Parser).literal_node,

		// extract declaration
		(*Parser).decl_node,

		// extract condition
		(*Parser).condition_node,

		// extract binary operators
		(*Parser).binary_operator_node,
		(*Parser).identifier,
		//fallback
		(*Parser).undefined_node,
	)
}

// todo: this might be only correct for lisp, other languages can be split by newline/semicolon
func (p *Parser) expression_node() *Node {
	if p.eq(p.Lexic.Punctuators.Open) {
		p.eat()
		nodes := []*Node{}

		for !p.eq(p.Lexic.Punctuators.Close) {
			// expression consists of atoms which can be other expressions, strings or basic atoms
			node := p.atom_node()
			log.Println("discovered atom within expression", node)
			if node != nil {
				nodes = append(nodes, node)
			}
		}
		p.eat()
		return node(Expression, []byte{}, first_known_type_or_unknown(nodes...), nodes)
	}
	return nil
}

func (p Parser) setup_argument_references(node *Node, arguments *Node) {
	for _, n := range node.Nodes {
		for i, a := range arguments.Nodes {
			// the argument name corresponds the atom
			if slices.Equal(a.Value, n.Value) {
				n.Kind = ArgumentVariable
				log.Println("argument reference setup for", a, n)
				n.ArgumentIndex = i
				if a.Type == Unknown && n.Type != Unknown {
					a.Type = n.Type
				} else {
					if a.Type != n.Type {
						p.record_error(fmt.Errorf("Inconsistent typing of argument %v", string(a.Value)))
					}
				}
			}
		}
		// going recursive: the arguments can be found in the enclosed expressions
		if len(n.Nodes) > 0 {
			p.setup_argument_references(n, arguments)
		}
	}
}

func (p *Parser) literal_node() *Node {
	if p.peek().TokenKind == LiteralBool {
		return node(Atom, p.eat().Value, Boolean, []*Node{})
	}
	if p.peek().TokenKind == LiteralString {
		return node(Atom, p.eat().Value, String, []*Node{})
	}
	if p.peek().TokenKind == LiteralInt {
		return node(Atom, p.eat().Value, Int, []*Node{})
	}
	return nil
}

func (p *Parser) decl_node() *Node {
	if p.peek().TokenKind == Keyword && p.eq("defun") {
		log.Println("function declaration: discard", p.eat())
		// we omit defun token, there is no value for node for it, we extract the callable symbol
		// SymbolAtom
		//  - Expression (arguments)
		//  - Body (arguments)
		name := p.eat()
		// extract arguments
		arguments := p.expression_node()

		log.Println("extracted arguments:", arguments)
		if arguments != nil {
			args := []string{}
			for _, arg := range arguments.Nodes {
				args = append(args, string(arg.Value))
			}
		}

		// extract body
		body := p.expression_node()
		log.Println("extracted body:", body)

		// reference beteween the nodes inside of the body to the arguments
		p.setup_argument_references(body, arguments)
		// currently the function gets redeclared globally but we want to support functions scope
		// e.g. function shadowed within function
		(*p.SymbolTable.Functions)[string(name.Value)] = FunctionEntry{Name: string(name.Value), Body: body, Arguments: arguments}
		return node(Declaration, name.Value, body.Type, []*Node{arguments, body})
	}
	return nil
}
func (p *Parser) condition_node() *Node {
	if p.peek().TokenKind == Keyword && p.eq("if") {
		name := p.eat().Value
		body := p.atom_node()
		left := p.atom_node()
		right := p.atom_node()
		t := first_known_type_or_unknown(left, right)
		return node(ConditionAtom, name, t, []*Node{body, left, right})

	}
	return nil
}

func (p *Parser) binary_operator_node() *Node {
	if p.peek().TokenKind == Operator {
		operator := p.eat()
		operand1 := p.atom_node()
		operand2 := p.atom_node()
		t := first_known_type_or_unknown(operand1, operand2)
		// TODO: this place looks hacky, need to figure out better way
		operand1.Type = t
		operand2.Type = t
		return node(BinaryOperator, operator.Value, t, []*Node{operand1, operand2})
	}
	return nil
}

func (p *Parser) identifier() *Node {
	if p.peek().TokenKind == Identifier {
		// variable or function call, resolved at runtime
		return node(SymbolAtom, p.eat().Value, Unknown, []*Node{})
	}
	return nil
}

func (p *Parser) peek() *Token {
	if !p.in() {
		return &Token{}
	}
	return p.Input[p.Position]
}

func (p *Parser) undefined_node() *Node {
	return node(Atom, p.eat().Value, Unknown, []*Node{})
}

func (p Parser) record_error(err error) {
	*p.Errors = append(*p.Errors, err)
}

func (p *Parser) in() bool {
	return p.Position < len(p.Input)
}

func (p *Parser) eat() *Token {
	p.Position += 1
	return p.Input[p.Position-1]
}

func (p *Parser) eq(str string) bool {
	for i, ch := range p.peek().Value {
		if str[i] != ch {
			return false
		}
	}
	return true
}

func first_known_type_or_unknown(input ...*Node) ResultType {
	for _, item := range input {
		if item != nil && item.Type != Unknown {
			return item.Type
		}
	}
	return Unknown
}
