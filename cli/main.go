package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	interpreter "ykrupin.com/lisp-interpreter-core"

	_ "net/http/pprof"
)

var lisp = interpreter.Lisp()
var maxStack = 100

func main() {
	//-e input expression
	//-i input file
	//-o output file
	expresion, oke := interpreter.GetArgument("-e", "--expression")
	input, oki := interpreter.GetArgument("-i", "--input")
	output, oko := interpreter.GetArgument("-o", "--output")
	ms, okm := interpreter.GetArgument("-ms", "--max-stack")
	profile, okp := interpreter.GetArgument("-p", "--profile")
	if okp {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		f, err := os.Create(profile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if okm {
		if msi, err := strconv.Atoi(ms); err != nil {
			log.Fatalln(err)
		} else {
			maxStack = msi
		}
	}

	if oke && oki {
		log.Fatalln("Error: Only one of two arguments should be supplied in the same time: -e/--expression or -i/--input")
	}
	if oko {
		defer interpreter.ToDisk(output, "log.log").Close()
	} else {
		log.SetOutput(io.Discard)
	}
	if oki {
		file(input)
	} else if oke {
		oneshot(expresion)
	} else {
		repl()
	}
}

func oneshot(expresion string) {
	runtime := interpreter.NewRuntime(maxStack)
	expressions := strings.Split(expresion, "\\n")
	log.Println("expression provided", expressions)
	for _, ex := range expressions {
		ast, _, _ := interpreter.ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
		r := runtime.EvaluateAst(ast)
		log.Println(r.String())
	}
}

func file(input string) {
	if b, err := os.ReadFile(input); err == nil {
		runtime := interpreter.NewRuntime(maxStack)
		str := string(b)
		expressions := strings.Split(str, "\n")
		log.Println("file input provided", input, "content:", strings.Join(expressions, ","))
		for _, ex := range expressions {
			ast, _, _ := interpreter.ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
			r := runtime.EvaluateAst(ast)
			log.Println(r.String())
		}
	} else {
		log.Fatal(err)
	}
}

func repl() {
	fmt.Println("Repl mode")
	runtime := interpreter.NewRuntime(maxStack)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sentence, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
		} else {
			ast, _, _ := interpreter.ParseAst(string(sentence[:len(sentence)-1]), runtime.SymbolTable, lisp)
			log.Println(ast.String())
			r := runtime.EvaluateAst(ast)
			fmt.Println(r.String())
		}
	}
}
