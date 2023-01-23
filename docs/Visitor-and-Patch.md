# Visitor and Patch

The [ast](https://pkg.go.dev/github.com/ilius/expr/ast?tab=doc) package 
provides `ast.Visitor` interface and `ast.Walk` function. You can use it for 
customizing of the AST before compiling.

For example, if you want to collect all variable names:

```go
package main

import (
	"fmt"

	"github.com/ilius/expr/ast"
	"github.com/ilius/expr/parser"
)

type visitor struct {
	identifiers []string
}

func (v *visitor) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		v.identifiers = append(v.identifiers, n.Value)
	}
}

func main() {
	tree, err := parser.Parse("foo + bar")
	if err != nil {
		panic(err)
	}

	visitor := &visitor{}
	ast.Walk(&tree.Node, visitor)

	fmt.Printf("%v", visitor.identifiers) // outputs [foo bar]
}
```

## Patch

Specify a visitor to modify the AST with `expr.Patch` function.  

```go
program, err := expr.Compile(code, expr.Patch(&visitor{}))
```
 
For example, we are going to replace expressions `list[-1]` with 
`list[len(list)-1]`.

```go
package main

import (
	"fmt"

	"github.com/ilius/expr"
	"github.com/ilius/expr/ast"
)

func main() {
	env := map[string]interface{}{
		"list": []int{1, 2, 3},
	}

	code := `list[-1]` // will output 3

	program, err := expr.Compile(code, expr.Env(env), expr.Patch(&patcher{}))
	if err != nil {
		panic(err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		panic(err)
	}
	fmt.Print(output)
}

type patcher struct{}

func (p *patcher) Visit(node *ast.Node) {
	n, ok := (*node).(*ast.IndexNode)
	if !ok {
		return
	}
	unary, ok := n.Index.(*ast.UnaryNode)
	if !ok {
		return
	}
	if unary.Operator == "-" {
		ast.Patch(&n.Index, &ast.BinaryNode{
			Operator: "-",
			Left:     &ast.BuiltinNode{Name: "len", Arguments: []ast.Node{n.Node}},
			Right:    unary.Node,
		})
	}

}
```

Type information is also available. Here is an example, there all `fmt.Stringer` 
interface automatically converted to `string` type.

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/ilius/expr"
	"github.com/ilius/expr/ast"
)

func main() {
	code := `Price == "$100"`

	program, err := expr.Compile(code, expr.Env(Env{}), expr.Patch(&stringerPatcher{}))
	if err != nil {
		panic(err)
	}

	env := Env{100_00}

	output, err := expr.Run(program, env)
	if err != nil {
		panic(err)
	}
	fmt.Print(output)
}

type Env struct {
	Price Price
}

type Price int

func (p Price) String() string {
	return fmt.Sprintf("$%v", int(p)/100)
}

var stringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

type stringerPatcher struct{}

func (p *stringerPatcher) Visit(node *ast.Node) {
	t := (*node).Type()
	if t == nil {
		return
	}
	if t.Implements(stringer) {
		ast.Patch(node, &ast.MethodNode{
			Node:   *node,
			Method: "String",
		})
	}

}
```

* Next: [Internals](Internals.md)
