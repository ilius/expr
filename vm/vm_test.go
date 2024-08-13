package vm_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ilius/expr/ast"
	"github.com/ilius/expr/checker"
	"github.com/ilius/expr/compiler"
	"github.com/ilius/expr/conf"
	"github.com/ilius/expr/parser"
	"github.com/ilius/expr/vm"
	"github.com/ilius/is/v2"
)

func TestRun_NilProgram(t *testing.T) {
	is := is.New(t)
	_, err := vm.Run(nil, nil)
	is.Err(err)
}

func TestRun_Debugger(t *testing.T) {
	is := is.New(t)
	input := `[1, 2]`

	node, err := parser.Parse(input)
	is.NotErr(err)

	program, err := compiler.Compile(node, nil)
	is.NotErr(err)

	debug := vm.Debug()
	go func() {
		debug.Step()
		debug.Step()
		debug.Step()
		debug.Step()
	}()
	go func() {
		for range debug.Position() {
		}
	}()

	_, err = debug.Run(program, nil)
	is.NotErr(err)
	is.Equal(len(debug.Stack()), 0)
	is.Nil(debug.Scope())
}

func TestRun_ReuseVM(t *testing.T) {
	is := is.New(t)
	node, err := parser.Parse(`map(1..2, {#})`)
	is.NotErr(err)

	program, err := compiler.Compile(node, nil)
	is.NotErr(err)

	reuse := vm.VM{}
	_, err = reuse.Run(program, nil)
	is.NotErr(err)
	_, err = reuse.Run(program, nil)
	is.NotErr(err)
}

func TestRun_Cast(t *testing.T) {
	is := is.New(t)
	input := `1`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	program, err := compiler.Compile(tree, &conf.Config{Expect: reflect.Float64})
	is.NotErr(err)

	out, err := vm.Run(program, nil)
	is.NotErr(err)
	is.Equal(float64(1), out)
}

func TestRun_Helpers(t *testing.T) {
	values := []interface{}{
		uint(1),
		uint8(1),
		uint16(1),
		uint32(1),
		uint64(1),
		1,
		int8(1),
		int16(1),
		int32(1),
		int64(1),
		float32(1),
		float64(1),
	}
	ops := []string{"+", "-", "*", "/", "%", "==", ">=", "<=", "<", ">"}

	for _, a := range values {
		for _, b := range values {
			for _, op := range ops {
				is := is.New(t)

				if op == "%" {
					switch a.(type) {
					case float32, float64:
						continue
					}
					switch b.(type) {
					case float32, float64:
						continue
					}
				}

				input := fmt.Sprintf("a %v b", op)
				env := map[string]interface{}{
					"a": a,
					"b": b,
				}

				config := conf.CreateNew()

				tree, err := parser.Parse(input)
				is.NotErr(err)

				_, err = checker.Check(tree, config)
				is.NotErr(err)

				program, err := compiler.Compile(tree, config)
				is.NotErr(err)

				_, err = vm.Run(program, env)
				is.NotErr(err)
			}
		}
	}
}

func TestRun_Helpers_Time(t *testing.T) {
	is := is.New(t)
	testTime := time.Date(2000, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	testDuration := time.Duration(1)

	tests := []struct {
		a       interface{}
		b       interface{}
		op      string
		want    interface{}
		wantErr bool
	}{
		{a: testTime, b: testTime, op: "<", wantErr: false, want: false},
		{a: testTime, b: testTime, op: ">", wantErr: false, want: false},
		{a: testTime, b: testTime, op: "<=", wantErr: false, want: true},
		{a: testTime, b: testTime, op: ">=", wantErr: false, want: true},
		{a: testTime, b: testTime, op: "==", wantErr: false, want: true},
		{a: testTime, b: testTime, op: "!=", wantErr: false, want: false},
		{a: testTime, b: testTime, op: "-", wantErr: false},
		{a: testTime, b: testDuration, op: "+", wantErr: false},

		// error cases
		{a: testTime, b: int64(1), op: "<", wantErr: true},
		{a: testTime, b: float64(1), op: "<", wantErr: true},
		{a: testTime, b: testDuration, op: "<", wantErr: true},

		{a: testTime, b: int64(1), op: ">", wantErr: true},
		{a: testTime, b: float64(1), op: ">", wantErr: true},
		{a: testTime, b: testDuration, op: ">", wantErr: true},

		{a: testTime, b: int64(1), op: "<=", wantErr: true},
		{a: testTime, b: float64(1), op: "<=", wantErr: true},
		{a: testTime, b: testDuration, op: "<=", wantErr: true},

		{a: testTime, b: int64(1), op: ">=", wantErr: true},
		{a: testTime, b: float64(1), op: ">=", wantErr: true},
		{a: testTime, b: testDuration, op: ">=", wantErr: true},

		{a: testTime, b: int64(1), op: "==", wantErr: false, want: false},
		{a: testTime, b: float64(1), op: "==", wantErr: false, want: false},
		{a: testTime, b: testDuration, op: "==", wantErr: false, want: false},

		{a: testTime, b: int64(1), op: "!=", wantErr: false, want: true},
		{a: testTime, b: float64(1), op: "!=", wantErr: false, want: true},
		{a: testTime, b: testDuration, op: "!=", wantErr: false, want: true},

		{a: testTime, b: int64(1), op: "-", wantErr: true},
		{a: testTime, b: float64(1), op: "-", wantErr: true},
		{a: testTime, b: testDuration, op: "-", wantErr: true},

		{a: testTime, b: testTime, op: "+", wantErr: true},
		{a: testTime, b: int64(1), op: "+", wantErr: true},
		{a: testTime, b: float64(1), op: "+", wantErr: true},
		{a: testDuration, b: testTime, op: "+", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("time helper test `%T %s %T`", tt.a, tt.op, tt.b), func(t *testing.T) {
			input := fmt.Sprintf("a %v b", tt.op)
			env := map[string]interface{}{
				"a": tt.a,
				"b": tt.b,
			}

			config := conf.CreateNew()

			tree, err := parser.Parse(input)
			is.NotErr(err)

			_, err = checker.Check(tree, config)
			is.NotErr(err)

			program, err := compiler.Compile(tree, config)
			is.NotErr(err)

			got, err := vm.Run(program, env)
			if tt.wantErr {
				is.Err(err)
			} else {
				is.NotErr(err)

				if tt.want != nil {
					is.Equal(tt.want, got)
				}
			}
		})
	}
}

func TestRun_MemoryBudget(t *testing.T) {
	is := is.New(t)
	input := `map(1..100, {map(1..100, {map(1..100, {0})})})`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	program, err := compiler.Compile(tree, nil)
	is.NotErr(err)

	_, err = vm.Run(program, nil)
	is.Err(err)
}

type ErrorEnv struct {
	InnerEnv InnerEnv
}
type InnerEnv struct{}

func (ErrorEnv) WillError(param string) (bool, error) {
	if param == "yes" {
		return false, errors.New("error")
	}
	return true, nil
}

func (ErrorEnv) FastError(...interface{}) interface{} {
	return true
}

func (InnerEnv) WillError(param string) (bool, error) {
	if param == "yes" {
		return false, errors.New("inner error")
	}
	return true, nil
}

func TestRun_MethodWithError(t *testing.T) {
	is := is.New(t)
	input := `WillError("yes")`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	is.NotErr(err)

	program, err := compiler.Compile(tree, funcConf)
	is.NotErr(err)

	out, err := vm.Run(program, env)
	is.ErrMsg(err, "error (1:1)\n | WillError(\"yes\")\n | ^")
	is.Equal(nil, out)
}

func TestRun_FastMethods(t *testing.T) {
	is := is.New(t)
	input := `hello() + world()`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	env := map[string]interface{}{
		"hello": func(...interface{}) interface{} { return "hello " },
		"world": func(...interface{}) interface{} { return "world" },
	}
	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	is.NotErr(err)

	program, err := compiler.Compile(tree, funcConf)
	is.NotErr(err)

	out, err := vm.Run(program, env)
	is.NotErr(err)
	is.Equal("hello world", out)
}

func TestRun_FastMethodWithError(t *testing.T) {
	is := is.New(t)
	input := `FastError()`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	is.NotErr(err)
	is.Msg("method must be fast").True(tree.Node.(*ast.CallNode).Fast)

	program, err := compiler.Compile(tree, funcConf)
	is.NotErr(err)

	out, err := vm.Run(program, env)
	is.NotErr(err)
	is.Equal(true, out)
}

func TestRun_InnerMethodWithError(t *testing.T) {
	is := is.New(t)
	input := `InnerEnv.WillError("yes")`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	program, err := compiler.Compile(tree, funcConf)
	is.NotErr(err)

	out, err := vm.Run(program, env)
	is.ErrMsg(err, "inner error (1:10)\n | InnerEnv.WillError(\"yes\")\n | .........^")
	is.Equal(nil, out)
}

func TestRun_InnerMethodWithError_NilSafe(t *testing.T) {
	is := is.New(t)
	input := `InnerEnv?.WillError("yes")`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	env := ErrorEnv{}
	funcConf := conf.New(env)
	program, err := compiler.Compile(tree, funcConf)
	is.NotErr(err)

	out, err := vm.Run(program, env)
	is.ErrMsg(err, "inner error (1:11)\n | InnerEnv?.WillError(\"yes\")\n | ..........^")
	is.Equal(nil, out)
}

func TestRun_TaggedFieldName(t *testing.T) {
	is := is.New(t)
	input := `value`

	tree, err := parser.Parse(input)
	is.NotErr(err)

	env := struct {
		V string `expr:"value"`
	}{
		V: "hello world",
	}

	funcConf := conf.New(env)
	_, err = checker.Check(tree, funcConf)
	is.NotErr(err)

	program, err := compiler.Compile(tree, funcConf)
	is.NotErr(err)

	out, err := vm.Run(program, env)
	is.NotErr(err)
	is.Equal("hello world", out)
}
