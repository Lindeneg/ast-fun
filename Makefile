.PHONY: repl
repl:
	go run main.go

.PHONY: file
file:
	go run main.go main.monkey


.PHONY: test
test:
	go test ./lexer ./parser ./ast ./evaluator

