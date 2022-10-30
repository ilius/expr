package optimizer_test

import (
	"strings"
	"testing"

	"github.com/ilius/expr/ast"
	"github.com/ilius/expr/checker"
	"github.com/ilius/expr/conf"
	"github.com/ilius/expr/optimizer"
	"github.com/ilius/expr/parser"
	"github.com/ilius/is/v2"
)

func TestOptimize_constant_folding(t *testing.T) {
	is := is.New(t)
	tree, err := parser.Parse(`[1,2,3][5*5-25]`)
	is.NotErr(err)

	err = optimizer.Optimize(&tree.Node, nil)
	is.NotErr(err)

	expected := &ast.MemberNode{
		Node:     &ast.ConstantNode{Value: []int{1, 2, 3}},
		Property: &ast.IntegerNode{Value: 0},
	}
	is.Equal(ast.Dump(expected), ast.Dump(tree.Node))
}

func TestOptimize_in_array(t *testing.T) {
	is := is.New(t)
	config := conf.New(map[string]int{"v": 0})

	tree, err := parser.Parse(`v in [1,2,3]`)
	is.NotErr(err)

	_, err = checker.Check(tree, config)
	is.NotErr(err)

	err = optimizer.Optimize(&tree.Node, nil)
	is.NotErr(err)

	expected := &ast.BinaryNode{
		Operator: "in",
		Left:     &ast.IdentifierNode{Value: "v"},
		Right:    &ast.ConstantNode{Value: map[int]struct{}{1: {}, 2: {}, 3: {}}},
	}
	is.Equal(ast.Dump(expected), ast.Dump(tree.Node))
}

func TestOptimize_in_range(t *testing.T) {
	is := is.New(t)
	tree, err := parser.Parse(`age in 18..31`)
	is.NotErr(err)

	err = optimizer.Optimize(&tree.Node, nil)
	is.NotErr(err)

	left := &ast.IdentifierNode{
		Value: "age",
	}
	expected := &ast.BinaryNode{
		Operator: "and",
		Left: &ast.BinaryNode{
			Operator: ">=",
			Left:     left,
			Right: &ast.IntegerNode{
				Value: 18,
			},
		},
		Right: &ast.BinaryNode{
			Operator: "<=",
			Left:     left,
			Right: &ast.IntegerNode{
				Value: 31,
			},
		},
	}
	is.Equal(ast.Dump(expected), ast.Dump(tree.Node))
}

func TestOptimize_const_range(t *testing.T) {
	is := is.New(t)
	tree, err := parser.Parse(`-1..1`)
	is.NotErr(err)

	err = optimizer.Optimize(&tree.Node, nil)
	is.NotErr(err)

	expected := &ast.ConstantNode{
		Value: []int{-1, 0, 1},
	}
	is.Equal(ast.Dump(expected), ast.Dump(tree.Node))
}

func TestOptimize_const_expr(t *testing.T) {
	is := is.New(t)
	tree, err := parser.Parse(`upper("hello")`)
	is.NotErr(err)

	env := map[string]interface{}{
		"upper": strings.ToUpper,
	}

	config := conf.New(env)
	config.ConstExpr("upper")

	err = optimizer.Optimize(&tree.Node, config)
	is.NotErr(err)

	expected := &ast.ConstantNode{
		Value: "HELLO",
	}
	is.Equal(ast.Dump(expected), ast.Dump(tree.Node))
}
