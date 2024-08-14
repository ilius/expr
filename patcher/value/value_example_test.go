package value_test

import (
	"fmt"

	"github.com/ilius/expr"
	"github.com/ilius/expr/patcher/value"
	"github.com/ilius/expr/vm"
)

type myInt struct {
	Int int
}

func (v *myInt) AsInt() int {
	return v.Int
}

func (v *myInt) AsAny() any {
	return v.Int
}

func ExampleAnyValuer() {
	env := make(map[string]any)
	env["ValueOne"] = &myInt{1}
	env["ValueTwo"] = &myInt{2}

	program, err := expr.Compile("ValueOne + ValueTwo", expr.Env(env), value.ValueGetter)

	if err != nil {
		panic(err)
	}

	out, err := vm.Run(program, env)

	if err != nil {
		panic(err)
	}

	fmt.Println(out)
	// Output: 3
}
