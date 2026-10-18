go -C ../core test \
&& go  -C ../core build  \
&& go build \
&& ./lisp-interpreter-cli -o build/test/work
#&& ./lisp-interpreter-cli -o build/test/work -e "(defun fib (n)  (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))\n (fib 2)" && ./lisp-interpreter-cli -o build/test/work -i "example/example-easy.lisp" 
