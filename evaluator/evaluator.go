package evaluator

import (
	"fmt"

	"github.com/lindeneg/monkey/ast"
	"github.com/lindeneg/monkey/object"
	"github.com/lindeneg/monkey/token"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(n ast.Node, e *object.Environment) object.Object {
	switch node := n.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, e)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, e)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.LetStatement:
		val := Eval(node.Value, e)
		if isError(val) {
			return val
		}
		e.Set(node.Name.Value, val)
	case *ast.CallExpression:
		function := Eval(node.Function, e)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, e)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: e, Body: body}
	case *ast.IfExpression:
		return evalIfExpression(node, e)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, e)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.Identifier:
		return evalIdentifier(node, e)
	case *ast.PrefixExpression:
		right := Eval(node.Right, e)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, e)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, e)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, e)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	}
	return nil
}

func evalProgram(stms []ast.Statement, e *object.Environment) object.Object {
	var result object.Object
	for _, stm := range stms {
		result = Eval(stm, e)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(stms []ast.Statement, e *object.Environment) object.Object {
	var result object.Object
	for _, stm := range stms {
		result = Eval(stm, e)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaled := Eval(e, env)
		if isError(evaled) {
			return []object.Object{evaled}
		}
		result = append(result, evaled)
	}
	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return val
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}
	extendEnv := extendFunctionEnv(function, args)
	evaled := Eval(function.Body, extendEnv)
	return unwrapReturnValue(evaled)
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for i, p := range fn.Parameters {
		env.Set(p.Value, args[i])
	}
	return env
}

func unwrapReturnValue(evaled object.Object) object.Object {
	if returnValue, ok := evaled.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return evaled
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case token.BANG:
		return evalBangOperatorExpression(right)
	case token.MINUS:
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case operator == token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case operator == token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case token.PLUS:
		return &object.Integer{Value: leftVal + rightVal}
	case token.MINUS:
		return &object.Integer{Value: leftVal - rightVal}
	case token.ASTERISK:
		return &object.Integer{Value: leftVal * rightVal}
	case token.SLASH:
		return &object.Integer{Value: leftVal / rightVal}
	case token.LT:
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case token.LT_OR_EQ:
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case token.GT:
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case token.GT_OR_EQ:
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case token.EQ:
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE // TODO null is truthy?
	default:
		if n, ok := right.(*object.Integer); ok && n.Value == 0 {
			return TRUE
		}
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIfExpression(ie *ast.IfExpression, e *object.Environment) object.Object {
	cond := Eval(ie.Condition, e)
	if isError(cond) {
		return cond
	}
	if isTruthy(cond) {
		return Eval(ie.Consequence, e)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, e)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		if n, ok := obj.(*object.Integer); ok && n.Value == 0 {
			return false
		}
		return true
	}
}

func isError(obj object.Object) bool {
	if obj == nil {
		return false
	}
	return obj.Type() == object.ERROR_OBJ
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
