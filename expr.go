package expr

import (
	"fmt"
	"reflect"

	"github.com/ilius/expr/ast"
	"github.com/ilius/expr/checker"
	"github.com/ilius/expr/compiler"
	"github.com/ilius/expr/conf"
	"github.com/ilius/expr/file"
	"github.com/ilius/expr/optimizer"
	"github.com/ilius/expr/parser"
	"github.com/ilius/expr/vm"
)

// Option for configuring config.
type Option func(c *conf.Config)

// Eval parses, compiles and runs given input.
func Eval(input string, env interface{}) (interface{}, error) {
	if _, ok := env.(Option); ok {
		return nil, fmt.Errorf("misused expr.Eval: second argument (env) should be passed without expr.Env")
	}

	tree, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}

	program, err := compiler.Compile(tree, nil)
	if err != nil {
		return nil, err
	}

	output, err := vm.Run(program, env)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Env specifies expected input of env for type checks.
// If struct is passed, all fields will be treated as variables,
// as well as all fields of embedded structs and struct itself.
// If map is passed, all items will be treated as variables.
// Methods defined on this type will be available as functions.
func Env(env interface{}) Option {
	return func(c *conf.Config) {
		if _, ok := env.(map[string]interface{}); ok {
			c.MapEnv = true
		} else {
			if reflect.ValueOf(env).Kind() == reflect.Map {
				c.DefaultType = reflect.TypeOf(env).Elem()
			}
		}
		c.Strict = true
		c.Types = conf.CreateTypesTable(env)
		c.Env = env
	}
}

// AllowUndefinedVariables allows to use undefined variables inside expressions.
// This can be used with expr.Env option to partially define a few variables.
// Note what this option is only works in map environment are used, otherwise
// runtime.fetch will panic as there is no way to get missing field zero value.
func AllowUndefinedVariables() Option {
	return func(c *conf.Config) {
		c.Strict = false
	}
}

// Operator allows to replace a binary operator with a function.
func Operator(operator string, fn ...string) Option {
	return func(c *conf.Config) {
		c.Operators[operator] = append(c.Operators[operator], fn...)
	}
}

// ConstExpr defines func expression as constant. If all argument to this function is constants,
// then it can be replaced by result of this func call on compile step.
func ConstExpr(fn string) Option {
	return func(c *conf.Config) {
		c.ConstExpr(fn)
	}
}

// AsBool tells the compiler to expect boolean result.
func AsBool() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Bool
	}
}

// AsInt64 tells the compiler to expect int64 result.
func AsInt64() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Int64
	}
}

// AsFloat64 tells the compiler to expect float64 result.
func AsFloat64() Option {
	return func(c *conf.Config) {
		c.Expect = reflect.Float64
	}
}

// Optimize turns optimizations on or off.
func Optimize(b bool) Option {
	return func(c *conf.Config) {
		c.Optimize = b
	}
}

// Patch adds visitor to list of visitors what will be applied before compiling AST to bytecode.
func Patch(visitor ast.Visitor) Option {
	return func(c *conf.Config) {
		c.Visitors = append(c.Visitors, visitor)
	}
}

// Compile parses and compiles given input expression to bytecode program.
func Compile(input string, ops ...Option) (*vm.Program, error) {
	config := &conf.Config{
		Operators:    make(map[string][]string),
		ConstExprFns: make(map[string]reflect.Value),
		Optimize:     true,
	}

	for _, op := range ops {
		op(config)
	}

	if len(config.Operators) > 0 {
		config.Visitors = append(config.Visitors, &conf.OperatorPatcher{
			Operators: config.Operators,
			Types:     config.Types,
		})
	}

	if err := config.Check(); err != nil {
		return nil, err
	}

	tree, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}

	if len(config.Visitors) > 0 {
		for _, v := range config.Visitors {
			// We need to perform types check, because some visitors may rely on
			// types information available in the tree.
			_, _ = checker.Check(tree, config)
			ast.Walk(&tree.Node, v)
		}
		_, err = checker.Check(tree, config)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = checker.Check(tree, config)
		if err != nil {
			return nil, err
		}
	}

	if config.Optimize {
		err = optimizer.Optimize(&tree.Node, config)
		if err != nil {
			if fileError, ok := err.(*file.Error); ok {
				return nil, fileError.Bind(tree.Source)
			}
			return nil, err
		}
	}

	program, err := compiler.Compile(tree, config)
	if err != nil {
		return nil, err
	}

	return program, nil
}

// Run evaluates given bytecode program.
func Run(program *vm.Program, env interface{}) (interface{}, error) {
	return vm.Run(program, env)
}
