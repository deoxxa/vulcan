package exproute

import (
	"fmt"
	"github.com/mailgun/vulcan/location"
	"go/ast"
	"go/parser"
)

func parseExpression(in string, l location.Location) (matcher, error) {
	expr, err := parser.ParseExpr(in)
	if err != nil {
		return nil, err
	}

	var matcher matcher
	matcher = &constMatcher{location: l}
	var item interface{}

	ast.Inspect(expr, func(n ast.Node) bool {
		// If error condition has been triggered, stop inspecting.
		if err != nil {
			return false
		}
		switch x := n.(type) {
		case *ast.BasicLit:
			fmt.Printf("Literal: %s\n", x.Value)
			err = addFunctionArgument(item, x.Value)
		case *ast.CallExpr:
			fmt.Println("Call, creating func")
			item = &funcCall{}
		case *ast.Ident:
			fmt.Printf("Set function name: %s\n", x.Name)
			setFunctionName(item, x.Name)
		default:
			if x != nil {
				fmt.Printf("Unsupported %T", n)
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return createMatcher(matcher, item)
}

func addFunctionArgument(in interface{}, val string) error {
	call, ok := in.(*funcCall)
	if !ok {
		return fmt.Errorf("Literals are only allowed as part of function calls")
	}
	call.args = append(call.args, val)
	return nil
}

func setFunctionName(in interface{}, val string) error {
	call, ok := in.(*funcCall)
	if !ok {
		return fmt.Errorf("Only function calls are allowed")
	}
	call.name = val
	return nil
}

func createMatcher(currentMatcher matcher, in interface{}) (matcher, error) {
	fn, ok := in.(*funcCall)
	if !ok {
		return nil, fmt.Errorf("Expected fucntion, got %T", in)
	}
	switch fn.name {
	case TrieRouteFn:
		return makeTrieRouteMatcher(currentMatcher, fn.args)
	}
	return nil, fmt.Errorf("Unsupported method: %s", fn.name)
}

type funcCall struct {
	name string
	args []interface{}
}

func makeTrieRouteMatcher(matcher matcher, params []interface{}) (matcher, error) {
	if len(params) <= 0 {
		return nil, fmt.Errorf("%s accepts at least one argument - path to match", TrieRouteFn)
	}
	args, err := toStrings(params)
	if err != nil {
		return nil, err
	}

	// The first 0..n-1 arguments are considered to be request methods, e.g. (POST|GET|DELETE)
	if len(args) > 1 {
		matcher = &methodMatcher{methods: args[:len(args)-1], matcher: matcher}
	}

	t, err := parseTrie(args[len(args)-1], matcher)
	if err != nil {
		return nil, fmt.Errorf("%s - failed to parse path expression, %s", err)
	}
	return t, nil
}

func makeRegexpRouteMatcher(matcher matcher, params []interface{}) (matcher, error) {
	if len(params) <= 0 {
		return nil, fmt.Errorf("%s needs at least one argument - path to match", RegexpRouteFn)
	}
	args, err := toStrings(params)
	if err != nil {
		return nil, err
	}

	// The first 0..n-1 arguments are considered to be request methods, e.g. (POST|GET|DELETE)
	if len(args) > 1 {
		matcher = &methodMatcher{methods: args[:len(args)-1], matcher: matcher}
	}

	t, err := newRegexpMatcher(args[len(args)-1], mapRequestToUrl, matcher)
	if err != nil {
		return nil, fmt.Errorf("Error %s(%s) - %s", RegexpRouteFn, params, err)
	}
	return t, nil
}

func toStrings(in []interface{}) ([]string, error) {
	out := make([]string, len(in))
	for i, v := range in {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Expected string, got %T", v)
		}
		out[i] = s
	}
	return out, nil
}

const (
	TrieRouteFn   = "TrieRoute"
	RegexpRouteFn = "RegexpRoute"
)
