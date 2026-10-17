# A home-grown lisp interpreter and Repl interface written in go

Pet project, hand-written for fun and learning of runtimes, interpretors and compilers. 

## Supported scope

For more information and examples see [tests](./core/core_test.go).

### Types
Three types are supported `int`, `string`, `bool`. The types are inferred on runtime from literals and being propagated if possible.

### Operators
Supported operators: `>=`, `<=`, `>`, `<`, `=`, `+`, `-`, `*`, `/`.

### Expressions
Expressions are supported, e.g. `(< 1 2)` or `+ 1 2`.

### Conditional
The `if` conditional operator is supported e.g. `(if (= 2 2) 1 2)`.

### Functions
Function declarations e.g. `(defun doublen (n) (* n 2))` and calls are supported e.g. `(doublen 2)`. 
Recursion is supported. The stack depth is limited by default.
Arguments are checked at runtime. 
The declaration is global and not scoped. The function declared again will override existing declaration.

## CLI
A command line inferface to the interpreter. Repl by default. Accepts file via `-i` or `--input`, accepts plain lisp expression or code via `-e` or `-expression`.

## Server
Repl as a server, hosts static terminal-like page and connects to the interpreter via websocket. The server version has limited default stack depth.

Hosted on `fly.io`.