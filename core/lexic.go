package core

type Lexic struct {
	Keywords    []string
	Punctuators Punctuators
	Operators   []string
	Literals    Literals
	Comments    Comments
}

type Comments struct {
	Line []string
}

type Punctuators struct {
	All   []string
	Open  string
	Close string
}

type Literals struct {
	True   []string
	Null   []string
	String []string
}

func Lisp() *Lexic {
	return &Lexic{
		Punctuators: Punctuators{
			[]string{"(", ")"}, "(", ")",
		},
		Keywords: []string{"if", "defun"},
		// note: longest operators first
		Operators: []string{">=", "<=", ">", "<", "=", "+", "-", "*", "/"},
		Literals: Literals{
			True:   []string{"t", "T", "'t", "'T"},
			Null:   []string{"nil", "NIL", "'nil", "'NIL"},
			String: []string{"\""},
		},
		Comments: Comments{
			Line: []string{";;"},
		},
	}
}
