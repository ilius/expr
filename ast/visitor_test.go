package ast_test

import (
	"testing"

	"github.com/ilius/expr/ast"
	"github.com/ilius/is/v2"
)

type visitor struct {
	identifiers []string
}

func (v *visitor) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		v.identifiers = append(v.identifiers, n.Value)
	}
}

func TestWalk(t *testing.T) {
	is := is.New(t)
	var node ast.Node
	node = &ast.BinaryNode{
		Operator: "+",
		Left:     &ast.IdentifierNode{Value: "foo"},
		Right:    &ast.IdentifierNode{Value: "bar"},
	}

	visitor := &visitor{}
	ast.Walk(&node, visitor)
	is.Equal([]string{"foo", "bar"}, visitor.identifiers)
}

type patcher struct{}

func (p *patcher) Visit(node *ast.Node) {
	if _, ok := (*node).(*ast.IdentifierNode); ok {
		*node = &ast.NilNode{}
	}
}

func TestWalk_patch(t *testing.T) {
	is := is.New(t)
	var node ast.Node
	node = &ast.BinaryNode{
		Operator: "+",
		Left:     &ast.IdentifierNode{Value: "foo"},
		Right:    &ast.IdentifierNode{Value: "bar"},
	}

	patcher := &patcher{}
	ast.Walk(&node, patcher)
	is.EqualType(&ast.NilNode{}, node.(*ast.BinaryNode).Left)
	is.EqualType(&ast.NilNode{}, node.(*ast.BinaryNode).Right)
}
