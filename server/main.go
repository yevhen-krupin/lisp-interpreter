package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
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
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}
	version()
	http.HandleFunc("/repl", serve)
	log.Printf("serve /repl as ws://localhost:%v/repl", port)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Printf("serve static home / as https://localhost:%v/", port)
	log.Println("binding port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func version() {
	f, err := os.OpenFile("./static/version.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	ver := CompileInfo{}.Get()
	_, err = f.WriteString(fmt.Sprintf("%v.%v", ver.Version, ver.Commit[:8]))
	if err != nil {
		log.Fatal(err)
	}
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

type CompileInfo struct {
	Package    string
	GoVersion  string
	Commit     string
	CommitTime string
	Modified   bool
	Version    string
}

func (CompileInfo) Get() CompileInfo {
	out := CompileInfo{}

	z, ok := debug.ReadBuildInfo()
	if !ok {
		return out
	}

	out.GoVersion = z.GoVersion
	out.Package = z.Path
	out.Version = z.Main.Version
	for _, s := range z.Settings {
		switch s.Key {
		case "vcs.revision":
			out.Commit = s.Value
		case "vcs.time":
			out.CommitTime = s.Value
		case "vcs.modified":
			out.Modified = s.Value == "true"
		}
	}

	return out
}
