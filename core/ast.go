package core

type Parser struct {
	SymbolTable *SymbolTable
	Errors      *[]Error
	Position    int
	Input       []*Token
	Lexic       *Lexic
}

type SymbolTable struct {
	Functions *map[string]*FunctionEntry
}

func ParseAst(input string, symbolTable *SymbolTable, lexic *Lexic) (*Node, TokenStream, []Error) {
	data, errors := tokenize(input, lexic)
	p := Parser{symbolTable, &errors, 0, data, lexic}
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
		beginning := p.eat()
		nodes := []*Node{}

		for !p.eq(p.Lexic.Punctuators.Close) {
			// expression consists of atoms which can be other expressions, strings or basic atoms
			node := p.atom_node()
			//log.Println("discovered atom within expression", node)
			if node != nil {
				nodes = append(nodes, node)
			}
		}
		p.eat()
		t, v := consolidate_type(nodes...)
		if v.T == Failure {
			p.record_error(v, &beginning.Location)
		}
		return node(Expression, beginning, t, nodes)
	}
	return nil
}

func (p Parser) setup_argument_references(node *Node, arguments *Node, location *Location) {
	for _, n := range node.Nodes {
		for i, a := range arguments.Nodes {
			// the argument name corresponds the atom
			if a.Value.Equal(n.Value) {
				n.Kind = ArgumentVariable
				//log.Println("argument reference setup for", a, n)
				n.ArgumentIndex = i
				_, v := consolidate_type(n, a)
				if v.T == Failure {
					p.record_error(v, node.Location)
				}
			}
		}
		// going recursive: the arguments can be found in the enclosed expressions
		if len(n.Nodes) > 0 {
			p.setup_argument_references(n, arguments, location)
		}
	}
}

func (p *Parser) literal_node() *Node {
	if p.peek().TokenKind == LiteralBool {
		return node(Atom, p.eat(), Boolean, []*Node{})
	}
	if p.peek().TokenKind == LiteralString {
		return node(Atom, p.eat(), String, []*Node{})
	}
	if p.peek().TokenKind == LiteralInt {
		return node(Atom, p.eat(), Int, []*Node{})
	}
	return nil
}

func (p *Parser) decl_node() *Node {
	if p.peek().TokenKind == Keyword && p.eq("defun") {
		defun := p.eat()
		//log.Println("function declaration: discard")
		// we omit defun token, there is no value for node for it, we extract the callable symbol
		// SymbolAtom
		//  - Expression (arguments)
		//  - Body (arguments)
		name := p.eat()
		// extract arguments
		arguments := p.expression_node()

		//log.Println("extracted arguments:", arguments)
		if arguments != nil {
			args := []string{}
			for _, arg := range arguments.Nodes {
				args = append(args, arg.Value.ValueString)
			}
		}

		entry := &FunctionEntry{Body: nil, Arguments: arguments, EvaluatedArguments: make([]Value, len(arguments.Nodes))}
		(*p.SymbolTable.Functions)[name.Value.ValueString] = entry // use hash
		//log.Println("declared", name.Value.ValueStr)
		// extract body
		body := p.expression_node()
		entry.Body = body
		//log.Println("extracted body:", body)

		// reference beteween the nodes inside of the body to the arguments
		p.setup_argument_references(body, arguments, &defun.Location)
		// currently the function gets redeclared globally but we want to support functions scope
		// e.g. function shadowed within function
		return node(Declaration, name, body.Type, []*Node{arguments, body})
	}
	return nil
}
func (p *Parser) condition_node() *Node {
	if p.peek().TokenKind == Keyword && p.eq("if") {
		name := p.eat()
		body := p.atom_node()
		left := p.atom_node()
		right := p.atom_node()
		t, v := consolidate_type(left, right)
		if v.T == Failure {
			p.record_error(v, &name.Location)
		}
		return node(ConditionAtom, name, t, []*Node{body, left, right})

	}
	return nil
}

func (p *Parser) binary_operator_node() *Node {
	if p.peek().TokenKind == Operator {
		operator := p.eat()
		operand1 := p.atom_node()
		operand2 := p.atom_node()
		t, v := consolidate_type(operand1, operand2)
		if v.T == Failure {
			p.record_error(v, &operator.Location)
		}
		// TODO: this place looks hacky, need to figure out better way
		operand1.Type = t
		operand2.Type = t
		return node(BinaryOperator, operator, t, []*Node{operand1, operand2})
	}
	return nil
}

func (p *Parser) identifier() *Node {
	if p.peek().TokenKind == Identifier {
		name := p.eat()
		// todo: move to expression
		//log.Println("looking for function ", name.Value.ValueStr)
		// function call
		if name.Value.T == String {
			if f, ok := (*p.SymbolTable.Functions)[name.Value.ValueString]; ok {
				var args = make([]*Node, len(f.Arguments.Nodes))
				for i, a := range f.Arguments.Nodes {
					args[i] = p.atom_node()
					if args[i].Type != a.Type {
						p.record_error(TypeConsolidationError(args[i].Value, a.Type), args[i].Location)
					}
				}
				// check arguments
				if len(f.Arguments.Nodes) != len(args) {
					p.record_error(ArgumentsCountError(name.Value.ValueString, len(f.Arguments.Nodes), len(args)), &name.Location)
				}

				return funcnode(CallSite, name, Unknown, args, f)
			}
		}

		// variable
		return node(SymbolAtom, name, Unknown, []*Node{})

	}
	return nil
}

func (p *Parser) peek() *Token {
	if !p.in() {
		return &Token{TokenKind: EOF}
	}
	return p.Input[p.Position]
}

func (p *Parser) undefined_node() *Node {
	if p.in() {
		return node(Atom, p.eat(), Unknown, []*Node{})
	}
	return nil
}

func (p Parser) record_error(err Value, location *Location) {
	*p.Errors = append(*p.Errors, Error{err, *location})
}

func (p *Parser) in() bool {
	return p.Position < len(p.Input)
}

func (p *Parser) eat() *Token {
	p.Position += 1
	return p.Input[p.Position-1]
}

func (p *Parser) eq(str string) bool {
	return p.peek().Value.ValueString == str
}

func consolidate_type(input ...*Node) (ResultType, Value) {
	first := Unknown
	for _, item := range input {
		if first == Unknown && item.Type != Unknown {
			first = item.Type
		} else {
			if item.Type == Unknown {
				item.Type = first
			} else if item.Type != first {
				return Failure, TypeConsolidationError(item.Value, first)
			}
		}
	}
	return first, Value{} // empty value to show there is no error
}
