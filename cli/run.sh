# go -C ../core test -run TestRecursion \
# go -C ../core test && \
# go  -C ../core build && \
go build \
&& ./lisp-interpreter-cli  # pure repl
# && go  -C ../core build -gcflags '-m -l' \
# && ./lisp-interpreter-cli -e "(defun fib (n)  (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))\n (fib 35)" -p "profile.prof"
 #&& ./lisp-interpreter-cli -e "(defun doublen (n) (* n 2))\n(doublen 2)" -p "profile.prof"
## && ./lisp-interpreter-cli  --profile profile.prof
#&& ./lisp-interpreter-cli -o build/test/work -e "(defun fib (n)  (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))\n (fib 2)" && ./lisp-interpreter-cli -o build/test/work -i "example/example-easy.lisp" 

# useful profiling commands
# go tool pprof profile.prof   
    # top
    # list {name}
    # web
# go tool pprof -alloc_space http://localhost:6060/debug/pprof/heap
# go tool pprof -http=:8123 profile.prof   