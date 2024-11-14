.PHONY: repl
repl:
	go run main.go

.PHONY: file
file:
	go run main.go main.monkey


.PHONY: test
test:
	go test ./lexer ./parser ./ast ./evaluator

.PHONY: bdebug
bdebug:
	go build -gcflags=all="-N -l" -o ./bin/monkey main.go

.PHONY: debug
debug: debug
	gdb ./bin/monkey

