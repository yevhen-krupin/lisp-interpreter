package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	interpreter "ykrupin.com/lisp-interpreter-core"
)

var lisp = interpreter.Lisp()
var maxStack = 15

func main() {
	port, okp := interpreter.GetArgument("-p", "--port")
	ms, okm := interpreter.GetArgument("-ms", "--max-stack")
	if okm {
		if msp, err := strconv.Atoi(ms); err == nil {
			maxStack = msp
		} else {
			log.Fatal("unable to parse --max-stack or -ms value, the integer is expected but was:", ms)
		}
	}
	if !okp {
		port = "8080"
	}
	http.HandleFunc("/repl", serve)
	log.Printf("serve /repl as ws://localhost:%v/repl", port)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Printf("serve static home / as https://localhost:%v/", port)
	log.Println("binding port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serve(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	log.Println("trigger")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	runtime := interpreter.NewRuntime(maxStack)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		expressions := strings.Split(string(message), "\\n")
		log.Println("input provided", expressions)
		for _, ex := range expressions {
			ast, t, err := interpreter.ParseAst(strings.TrimLeft(ex, "\n"), runtime.SymbolTable, lisp)
			r := runtime.EvaluateAst(ast)
			resp := &Response{Tokens: t, Ast: ast, Result: r, ResultString: r.String(), Errors: err}
			if ser, err := json.Marshal(resp); err != nil {
				log.Println(err)
			} else {
				if err = c.WriteMessage(mt, ser); err != nil {
					log.Println(err)
				}
			}

		}
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

type Response struct {
	Tokens       interpreter.TokenStream      `json:"tokens"`
	Ast          *interpreter.Node            `json:"ast"`
	Result       interpreter.EvaluationResult `json:"result"`
	ResultString string                       `json:"resultString"`
	Errors       []error                      `json:"errors"`
}
