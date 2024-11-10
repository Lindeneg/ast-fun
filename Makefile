run:
	go run main.go

test:
	go test ./lexer ./parser ./ast ./evaluator
