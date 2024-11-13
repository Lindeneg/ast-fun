package evaluator

import (
	"fmt"
	"strings"

	"github.com/lindeneg/monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	"println": {
		Fn: func(args ...object.Object) object.Object {
			var sb strings.Builder
			for _, a := range args {
				switch arg := a.(type) {
				case *object.String:
					sb.WriteString(arg.Value)
				case *object.Integer:
					sb.WriteString(fmt.Sprintf("%d", arg.Value))
				default:
					return newError("argument to `println` not supported, got %s",
						args[0].Type())
				}
			}
			if sb.Len() > 0 {
				fmt.Println(sb.String())
			}
			return nil
		},
	},
}
