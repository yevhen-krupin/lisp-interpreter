package core

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"testing"
)

func TestTokenizer(tt *testing.T) {
	var cases []TestCase[[]*Token] = []TestCase[[]*Token]{
		{";;comment", t(c(";;comment", 0, 0))},
		// digits
		{"1", t(d(1, 0, 0))},
		{"32", t(d(32, 0, 0))},
		{"255", t(d(255, 0, 0))},
		{"256", t(d(256, 0, 0))},
		{"16777217", t(d(16777217, 0, 0))},
		// basic math, binary operations
		{"(+ 1 2)", t(so(0, 0), o("+", 0, 1), d(1, 0, 3), d(2, 0, 5), sc(0, 6))},
		{"(- 1 2)", t(so(0, 0), o("-", 0, 1), d(1, 0, 3), d(2, 0, 5), sc(0, 6))},
		{"(* 1 2)", t(so(0, 0), o("*", 0, 1), d(1, 0, 3), d(2, 0, 5), sc(0, 6))},
		{"(/ 1 2)", t(so(0, 0), o("/", 0, 1), d(1, 0, 3), d(2, 0, 5), sc(0, 6))},
		{"(/1 2)", t(so(0, 0), o("/", 0, 1), d(1, 0, 2), d(2, 0, 4), sc(0, 5))},
		{"(/ 1 2 )", t(so(0, 0), o("/", 0, 1), d(1, 0, 3), d(2, 0, 5), sc(0, 7))},
		{"( / 1 2)", t(so(0, 0), o("/", 0, 2), d(1, 0, 4), d(2, 0, 6), sc(0, 7))},
		// identifier, ending with whitespace or punctuation
		{"(if () 1 2)", t(so(0, 0), kw("if", 0, 1), so(0, 4), sc(0, 5), d(1, 0, 7), d(2, 0, 9), sc(0, 10))},
		{"(if() 1 2)", t(so(0, 0), kw("if", 0, 1), so(0, 3), sc(0, 4), d(1, 0, 6), d(2, 0, 8), sc(0, 9))},
		{"(n)", t(so(0, 0), i("n", 0, 1), sc(0, 2))},
		{"(n )", t(so(0, 0), i("n", 0, 1), sc(0, 3))},
		{"( n)", t(so(0, 0), i("n", 0, 2), sc(0, 3))},
		{"( n )", t(so(0, 0), i("n", 0, 2), sc(0, 4))},
		{"(n())", t(so(0, 0), i("n", 0, 1), so(0, 2), sc(0, 3), sc(0, 4))},
		{"(n ())", t(so(0, 0), i("n", 0, 1), so(0, 3), sc(0, 4), sc(0, 5))},
		// function difinitions
		{"(defun fact (n) (if (<= n 1) 1 (* n (fact (- n 1)))))",
			t(so(0, 0), f(0, 1), i("fact", 0, 7), so(0, 12), i("n", 0, 13), sc(0, 14), so(0, 16), kw("if", 0, 17), so(0, 20), o("<=", 0, 21), i("n", 0, 24), d(1, 0, 26), sc(0, 27), d(1, 0, 29), so(0, 31), o("*", 0, 32), i("n", 0, 34), so(0, 36), i("fact", 0, 37), so(0, 42), o("-", 0, 43), i("n", 0, 45), d(1, 0, 47), sc(0, 48), sc(0, 49), sc(0, 50), sc(0, 51), sc(0, 52))},
		{"(defun doublen (n) (* n 2))",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 15), i("n", 0, 16), sc(0, 17), so(0, 19), o("*", 0, 20), i("n", 0, 22), d(2, 0, 24), sc(0, 25), sc(0, 26))},
		{"(defun doublen(n) (* n 2))",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 14), i("n", 0, 15), sc(0, 16), so(0, 18), o("*", 0, 19), i("n", 0, 21), d(2, 0, 23), sc(0, 24), sc(0, 25))},
		{"(defun doublen (n)(* n 2))",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 15), i("n", 0, 16), sc(0, 17), so(0, 18), o("*", 0, 19), i("n", 0, 21), d(2, 0, 23), sc(0, 24), sc(0, 25))},
		{"(defun doublen(n)(* n 2))",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 14), i("n", 0, 15), sc(0, 16), so(0, 17), o("*", 0, 18), i("n", 0, 20), d(2, 0, 22), sc(0, 23), sc(0, 24))},
		{"(defun doublen (n) (*n 2))",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 15), i("n", 0, 16), sc(0, 17), so(0, 19), o("*", 0, 20), i("n", 0, 21), d(2, 0, 23), sc(0, 24), sc(0, 25))},
		{"(defun doublen (n) (* n 2) )",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 15), i("n", 0, 16), sc(0, 17), so(0, 19), o("*", 0, 20), i("n", 0, 22), d(2, 0, 24), sc(0, 25), sc(0, 27))},
		//multiline
		{
			"(defun doublen (n) (* n 2))\n(doublen 2)",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 15), i("n", 0, 16), sc(0, 17), so(0, 19), o("*", 0, 20), i("n", 0, 22), d(2, 0, 24), sc(0, 25), sc(0, 26), so(1, 0), i("doublen", 1, 1), d(2, 1, 9), sc(1, 10)),
		},
		// same but with whitespace at the beginning of the second line
		{
			"(defun doublen (n) (* n 2))\n (doublen 2)",
			t(so(0, 0), f(0, 1), i("doublen", 0, 7), so(0, 15), i("n", 0, 16), sc(0, 17), so(0, 19), o("*", 0, 20), i("n", 0, 22), d(2, 0, 24), sc(0, 25), sc(0, 26), so(1, 1), i("doublen", 1, 2), d(2, 1, 10), sc(1, 11)),
		},
		// strings
		{"\"Hello, Coding Challenges\"",
			t(s("Hello, Coding Challenges", 0, 0))},
		{"(defun hello() (\"Hello Coding Challenges\"))\n(hello)",
			t(so(0, 0), f(0, 1), i("hello", 0, 7), so(0, 12), sc(0, 13), so(0, 15), s("Hello Coding Challenges", 0, 16), sc(0, 41), sc(0, 42), so(1, 0), i("hello", 1, 1), sc(1, 6))},
	}

	for index, element := range cases {
		tt.Run(fmt.Sprintf("tokenizer_test [%d] %v", index, element.input), func(t *testing.T) {

			file := fmt.Sprintf("%v_%v.log", "tokenizer_test", index)
			s := ToDisk(TEST_DIR, file)
			defer s.Close()
			tokens := slices.Collect(tokenize(element.input, Lisp()))
			if !eq(tokens, element.output) {
				t.Errorf("The input '%v' was expected to be tokenized as '%v' but was tokenized as '%v', log: %v", element.input, TokenStream(element.output).Diagnostic(), TokenStream(tokens).Diagnostic(), s.f.Name())
			}
		})
	}
}

func TestEvaluation(tt *testing.T) {
	var strings = []TestCase[string]{
		{"\"Hello, Coding Challenges\"",
			"Hello, Coding Challenges"},
		{"(defun hello() (\"Hello Coding Challenges\"))\n(hello)",
			"Hello Coding Challenges"},
		// function declaration emit function name
		{"(defun fib (n)  (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))", "FIB"},
	}
	var bools = []TestCase[bool]{
		{"t", true},
		{"t", true},
		{"T", true},
		{"T", true},
		{"nil", false},
		{"nil", false},
		{"NIL", false},
		{"'NIL", false},
		{"()", false},
		{"(< 1 2)", true},
		{"(< 2 2)", false},
		{"(< 3 2)", false},
		{"(> 1 2)", false},
		{"(> 2 2)", false},
		{"(> 3 2)", true},
		{"(= 1 2)", false},
		{"(= 2 2)", true},
		{"(= 3 2)", false},
	}
	var ints = []TestCase[int]{
		{"1", 1},
		{"32", 32},
		{"255", 255},
		{"256", 256},
		{"65536", 65536},
		{"65537", 65537},
		{"16777216", 16777216},
		{"16777217", 16777217},
		{"(+ 1 2)", 3},
		{"(* 1 2)", 2},
		{"(* 5 3)", 15},
		{"(/ 5 5)", 1},
		{"(/ 6 3)", 2},
		{"(- 1 1)", 0},
		{"(- 1 2)", -1},
		{"- 1 2", -1},
		{"(defun doublen (n) (* n 2))\n (doublen 2)", 4},
		// conditional evaluation
		{"(if (= 2 2) 1 2)", 1},
		// () evaluates as false
		{"(if () 1 2)", 2},
		{"(defun twoorthree (n) (if (= n 2) 2 3))\n (twoorthree 2)", 2},
		{"(defun twoorthree (n) (if (= n 2) 2 3))\n (twoorthree 1)", 3},

		// redeclare should take latest version
		{"(defun twoorthree (n) (if (= n 2) 2 3))\n (defun twoorthree (n) (if (= n 2) 2 5))\n (twoorthree 1)", 5},

		// fibonacci
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 0)", 0},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 1)", 1},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 2)", 1},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 3)", 2},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 4)", 3},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 5)", 5},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 6)", 8},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 7)", 13},
		{"(defun fib (n)" +
			"  (if (< n 2)" +
			"      n" +
			"      (+ (fib (- n 1)) (fib (- n 2)))))" +
			"\n (fib 8)", 21},

		// comments
		{";; comment whatever\n5", 5},
		{";; comment whatever\n5;; comment whatever", 5},
	}

	for index, element := range strings {
		tt.Run(fmt.Sprintf("string test [%d] %v", index, element.input), func(t *testing.T) {
			result, file := evaluate_test_case("TestEvaluation_strings", index, element)
			if result.Value != element.output {
				t.Errorf("The input '%v' was expected to be evaluated as '%v' but was evaluated as '%v', log: %v", element.input, element.output, result.Value, file)
			}
		})
	}
	for index, element := range bools {
		tt.Run(fmt.Sprintf("bool test [%d] %v", index, element.input), func(t *testing.T) {
			result, file := evaluate_test_case("TestEvaluation_bools", index, element)
			if result.Value != element.output {
				t.Errorf("The input '%v' was expected to be evaluated as '%v' but was evaluated as '%v', log: %v", element.input, element.output, result.Value, file)
			}
		})
	}
	for index, element := range ints {
		tt.Run(fmt.Sprintf("int test [%d] %v", index, element.input), func(t *testing.T) {
			result, file := evaluate_test_case("TestEvaluation_ints", index, element)
			if result.Value != element.output {
				t.Errorf("The input '%v' was expected to be evaluated as '%v' but was evaluated as '%v', log: %v", element.input, element.output, result.Value, file)
			}
		})
	}

	/**

	array := []string{
		// ":CC",
		"(format t \"Hello, Coding Challenge World World\")",
	}

	for index, element := range array {
		runtime := NewRuntime()
		for sub_element := range strings.SplitSeq(element, "\n") {
			ast, _, _ := ParseAst(sub_element)
			log.Println("Expression:", index, ": `", sub_element, "`")
			r := runtime.EvaluateAst(ast)
			if r.Error != nil {
				log.Println("Failed to evaluate due to", r.Error.Error())
			} else {
				log.Println(r.String())
				log.Println(ast.String())
			}
		}
	}**/

}

func kw(s string, line int, pos int) *Token {
	return &Token{Keyword, []byte(s), Location{line, pos}}
}

func so(line int, pos int) *Token {
	return &Token{Punctuator, []byte{'('}, Location{line, pos}}
}

func sc(line int, pos int) *Token {
	return &Token{Punctuator, []byte{')'}, Location{line, pos}}
}

func o(op string, line int, pos int) *Token {
	return &Token{Operator, []byte(op), Location{line, pos}}
}

func t(args ...*Token) []*Token {
	return args
}

func d(i int, line int, pos int) *Token {
	return &Token{LiteralInt, int32_to_bytes(i), Location{line, pos}}
}

func f(line int, pos int) *Token {
	return &Token{Keyword, []byte("defun"), Location{line, pos}}
}

func i(s string, line int, pos int) *Token {
	return &Token{Identifier, []byte(s), Location{line, pos}}
}

func s(s string, line int, pos int) *Token {
	return &Token{LiteralString, []byte(s), Location{line, pos}}
}

func c(c string, line int, pos int) *Token {
	return &Token{Comment, []byte(c), Location{line, pos}}
}

func eq(a []*Token, b []*Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if e.TokenKind != b[i].TokenKind || e.Location.Line != b[i].Location.Line || e.Location.Position != b[i].Location.Position {
			return false
		}
		for j, ch := range e.Value {
			if ch != b[i].Value[j] {
				return false
			}
		}
	}

	return true
}

const TEST_DIR = "build/test/work"

type TestCase[T int | bool | string | []*Token] struct {
	input  string
	output T
}

type TestResult struct {
	pass int
	fail int
}

func evaluate_test_case[T int | bool | string](name string, index int, element TestCase[T]) (EvaluationResult, string) {
	results := []EvaluationResult{}
	nodes := []*Node{}
	subs := strings.SplitSeq(element.input, "\n")

	runtime := NewRuntime(10)
	file := fmt.Sprintf("%v_%v.log", name, index)
	s := ToDisk(TEST_DIR, file)
	defer s.Close()
	for sub_element := range subs {
		ast, tokens, _ := ParseAst(sub_element, runtime.SymbolTable, Lisp())
		log.Println(tokens.String())
		log.Println(ast.String())
		nodes = append(nodes, ast)
		r := runtime.EvaluateAst(ast)
		results = append(results, r)
		if r.Error != nil {
			log.Println("Failed to evaluate due to", r.Error.Error())
		}
	}
	final := results[len(results)-1]

	log.Println("Functions declared:", runtime.SymbolTable.Functions)
	for i, r := range results {
		if i == len(results)-1 {
			log.Println("Evaluated as", r.Type, r.Value, " but has to be", element.output)
		} else {
			log.Println("Evaluated as:", r.Type, r.Value)
		}
		log.Println(nodes[i].String())
	}

	return final, s.f.Name()
}
