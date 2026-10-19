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
		{";;comment", t(c(";;comment"))},
		{"1", t(d(1))},
		{"32", t(d(32))},
		{"255", t(d(255))},
		{"256", t(d(256))},
		{"16777217", t(d(16777217))},
		{"(+ 1 2)", t(so(), o("+"), d(1), d(2), sc())},
		{"(- 1 2)", t(so(), o("-"), d(1), d(2), sc())},
		{"(* 1 2)", t(so(), o("*"), d(1), d(2), sc())},
		{"(/ 1 2)", t(so(), o("/"), d(1), d(2), sc())},
		{"(/1 2)", t(so(), o("/"), d(1), d(2), sc())},
		{"(/ 1 2 )", t(so(), o("/"), d(1), d(2), sc())},
		{"( / 1 2)", t(so(), o("/"), d(1), d(2), sc())},
		{"(defun fact (n) (if (<= n 1) 1 (* n (fact (- n 1)))))", t(so(), f(), i("fact"), so(), i("n"), sc(), so(), kw("if"), so(), o("<="), i("n"), d(1), sc(), d(1), so(), o("*"), i("n"), so(), i("fact"), so(), o("-"), i("n"), d(1), sc(), sc(), sc(), sc(), sc())},
		{"(defun doublen (n) (* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc())},
		{"(defun doublen(n) (* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc())},
		{"(defun doublen (n)(* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc())},
		{"(defun doublen(n)(* n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc())},
		{"(defun doublen (n) (*n 2))", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc())},
		{"(defun doublen (n) (* n 2)   )", t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc())},
		{
			"(defun doublen (n) (* n 2))\n (doublen 2)",
			t(so(), f(), i("doublen"), so(), i("n"), sc(), so(), o("*"), i("n"), d(2), sc(), sc(), so(), i("doublen"), d(2), sc()),
		},
		{"\"Hello, Coding Challenges\"",
			t(s("\"Hello, Coding Challenges\""))},
		{"(defun hello() (\"Hello Coding Challenges\"))\n(hello)", t(so(), f(), i("hello"), so(), sc(), so(), s("\"Hello Coding Challenges\""), sc(), sc(), so(), i("hello"), sc())},
	}

	for index, element := range cases {
		tt.Run(fmt.Sprintf("tokenizer_test [%d] %v", index, element.input), func(t *testing.T) {

			file := fmt.Sprintf("%v_%v.log", "tokenizer_test", index)
			s := ToDisk(TEST_DIR, file)
			defer s.Close()
			tokens := slices.Collect(tokenize(element.input, Lisp()))
			if !eq(tokens, element.output) {
				t.Errorf("The input '%v' was expected to be tokenized as '%v' but was tokenized as '%v', log: %v", element.input, TokenStream(element.output).String(), TokenStream(tokens).String(), s.f.Name())
			}
		})
	}
}

func TestEvaluation(tt *testing.T) {
	var strings = []TestCase[string]{
		{"\"Hello, Coding Challenges\"",
			"\"Hello, Coding Challenges\""},
		{"(defun hello() (\"Hello Coding Challenges\"))\n(hello)",
			"\"Hello Coding Challenges\""},
		// function declaration emit function name
		{"(defun fib (n)  (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))", "FIB"},
	}
	var bools = []TestCase[bool]{
		{"t", true},
		{"T", true},
		{"nil", false},
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
		"()",
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

func kw(s string) *Token {
	return &Token{Keyword, []byte(s)}
}

func so() *Token {
	return &Token{Punctuator, []byte{'('}}
}

func sc() *Token {
	return &Token{Punctuator, []byte{')'}}
}

func o(op string) *Token {
	return &Token{Operator, []byte(op)}
}

func t(args ...*Token) []*Token {
	return args
}

func d(i int) *Token {
	return &Token{Literal, int32_to_bytes(i)}
}

func f() *Token {
	return &Token{Keyword, []byte("defun")}
}

func i(s string) *Token {
	return &Token{Identifier, []byte(s)}
}

func s(s string) *Token {
	return &Token{Literal, []byte(s)}
}

func c(c string) *Token {
	return &Token{Comment, []byte(c)}
}

func eq(a []*Token, b []*Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if e.TokenKind != b[i].TokenKind {
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
		ast, _, _ := ParseAst(sub_element, runtime.SymbolTable, Lisp())
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
