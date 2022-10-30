package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ilius/expr/ast"
	"github.com/ilius/expr/parser"
	"github.com/ilius/is/v2"
)

func TestParse(t *testing.T) {
	parseTests := []struct {
		input    string
		expected ast.Node
	}{
		{
			"a",
			&ast.IdentifierNode{Value: "a"},
		},
		{
			`"str"`,
			&ast.StringNode{Value: "str"},
		},
		{
			"3",
			&ast.IntegerNode{Value: 3},
		},
		{
			"0xFF",
			&ast.IntegerNode{Value: 255},
		},
		{
			"10_000_000",
			&ast.IntegerNode{Value: 10_000_000},
		},
		{
			"2.5",
			&ast.FloatNode{Value: 2.5},
		},
		{
			"1e9",
			&ast.FloatNode{Value: 1e9},
		},
		{
			"true",
			&ast.BoolNode{Value: true},
		},
		{
			"false",
			&ast.BoolNode{Value: false},
		},
		{
			"nil",
			&ast.NilNode{},
		},
		{
			"-3",
			&ast.UnaryNode{Operator: "-", Node: &ast.IntegerNode{Value: 3}},
		},
		{
			"1 - 2",
			&ast.BinaryNode{Operator: "-", Left: &ast.IntegerNode{Value: 1}, Right: &ast.IntegerNode{Value: 2}},
		},
		{
			"(1 - 2) * 3",
			&ast.BinaryNode{
				Operator: "*",
				Left: &ast.BinaryNode{
					Operator: "-", Left: &ast.IntegerNode{Value: 1},
					Right: &ast.IntegerNode{Value: 2},
				}, Right: &ast.IntegerNode{Value: 3},
			},
		},
		{
			"a or b or c",
			&ast.BinaryNode{Operator: "or", Left: &ast.BinaryNode{Operator: "or", Left: &ast.IdentifierNode{Value: "a"}, Right: &ast.IdentifierNode{Value: "b"}}, Right: &ast.IdentifierNode{Value: "c"}},
		},
		{
			"a or b and c",
			&ast.BinaryNode{Operator: "or", Left: &ast.IdentifierNode{Value: "a"}, Right: &ast.BinaryNode{Operator: "and", Left: &ast.IdentifierNode{Value: "b"}, Right: &ast.IdentifierNode{Value: "c"}}},
		},
		{
			"(a or b) and c",
			&ast.BinaryNode{Operator: "and", Left: &ast.BinaryNode{Operator: "or", Left: &ast.IdentifierNode{Value: "a"}, Right: &ast.IdentifierNode{Value: "b"}}, Right: &ast.IdentifierNode{Value: "c"}},
		},
		{
			"2**4-1",
			&ast.BinaryNode{Operator: "-", Left: &ast.BinaryNode{Operator: "**", Left: &ast.IntegerNode{Value: 2}, Right: &ast.IntegerNode{Value: 4}}, Right: &ast.IntegerNode{Value: 1}},
		},
		{
			"foo(bar())",
			&ast.CallNode{Callee: &ast.IdentifierNode{Value: "foo"}, Arguments: []ast.Node{&ast.CallNode{Callee: &ast.IdentifierNode{Value: "bar"}, Arguments: []ast.Node{}}}},
		},
		{
			`foo("arg1", 2, true)`,
			&ast.CallNode{Callee: &ast.IdentifierNode{Value: "foo"}, Arguments: []ast.Node{&ast.StringNode{Value: "arg1"}, &ast.IntegerNode{Value: 2}, &ast.BoolNode{Value: true}}},
		},
		{
			"foo.bar",
			&ast.MemberNode{Node: &ast.IdentifierNode{Value: "foo"}, Property: &ast.StringNode{Value: "bar"}},
		},
		{
			"foo['all']",
			&ast.MemberNode{Node: &ast.IdentifierNode{Value: "foo"}, Property: &ast.StringNode{Value: "all"}},
		},
		{
			"foo.bar()",
			&ast.CallNode{Callee: &ast.MemberNode{Node: &ast.IdentifierNode{Value: "foo"}, Property: &ast.StringNode{Value: "bar"}}, Arguments: []ast.Node{}},
		},
		{
			`foo.bar("arg1", 2, true)`,
			&ast.CallNode{Callee: &ast.MemberNode{Node: &ast.IdentifierNode{Value: "foo"}, Property: &ast.StringNode{Value: "bar"}}, Arguments: []ast.Node{&ast.StringNode{Value: "arg1"}, &ast.IntegerNode{Value: 2}, &ast.BoolNode{Value: true}}},
		},
		{
			"foo[3]",
			&ast.MemberNode{Node: &ast.IdentifierNode{Value: "foo"}, Property: &ast.IntegerNode{Value: 3}},
		},
		{
			"true ? true : false",
			&ast.ConditionalNode{Cond: &ast.BoolNode{Value: true}, Exp1: &ast.BoolNode{Value: true}, Exp2: &ast.BoolNode{}},
		},
		{
			"a?[b]:c",
			&ast.ConditionalNode{Cond: &ast.IdentifierNode{Value: "a"}, Exp1: &ast.ArrayNode{Nodes: []ast.Node{&ast.IdentifierNode{Value: "b"}}}, Exp2: &ast.IdentifierNode{Value: "c"}},
		},
		{
			"a.b().c().d[33]",
			&ast.MemberNode{
				Node: &ast.MemberNode{
					Node: &ast.CallNode{
						Callee: &ast.MemberNode{
							Node: &ast.CallNode{
								Callee: &ast.MemberNode{
									Node: &ast.IdentifierNode{
										Value: "a",
									},
									Property: &ast.StringNode{
										Value: "b",
									},
								},
								Arguments: []ast.Node{},
								Fast:      false,
							},
							Property: &ast.StringNode{
								Value: "c",
							},
						},
						Arguments: []ast.Node{},
						Fast:      false,
					},
					Property: &ast.StringNode{
						Value: "d",
					},
				},
				Property: &ast.IntegerNode{
					Value: 33,
				},
			},
		},
		{
			"'a' == 'b'",
			&ast.BinaryNode{Operator: "==", Left: &ast.StringNode{Value: "a"}, Right: &ast.StringNode{Value: "b"}},
		},
		{
			"+0 != -0",
			&ast.BinaryNode{Operator: "!=", Left: &ast.UnaryNode{Operator: "+", Node: &ast.IntegerNode{}}, Right: &ast.UnaryNode{Operator: "-", Node: &ast.IntegerNode{}}},
		},
		{
			"[a, b, c]",
			&ast.ArrayNode{Nodes: []ast.Node{&ast.IdentifierNode{Value: "a"}, &ast.IdentifierNode{Value: "b"}, &ast.IdentifierNode{Value: "c"}}},
		},
		{
			"{foo:1, bar:2}",
			&ast.MapNode{Pairs: []ast.Node{&ast.PairNode{Key: &ast.StringNode{Value: "foo"}, Value: &ast.IntegerNode{Value: 1}}, &ast.PairNode{Key: &ast.StringNode{Value: "bar"}, Value: &ast.IntegerNode{Value: 2}}}},
		},
		{
			"{foo:1, bar:2, }",
			&ast.MapNode{Pairs: []ast.Node{&ast.PairNode{Key: &ast.StringNode{Value: "foo"}, Value: &ast.IntegerNode{Value: 1}}, &ast.PairNode{Key: &ast.StringNode{Value: "bar"}, Value: &ast.IntegerNode{Value: 2}}}},
		},
		{
			`{"a": 1, 'b': 2}`,
			&ast.MapNode{Pairs: []ast.Node{&ast.PairNode{Key: &ast.StringNode{Value: "a"}, Value: &ast.IntegerNode{Value: 1}}, &ast.PairNode{Key: &ast.StringNode{Value: "b"}, Value: &ast.IntegerNode{Value: 2}}}},
		},
		{
			"[1].foo",
			&ast.MemberNode{Node: &ast.ArrayNode{Nodes: []ast.Node{&ast.IntegerNode{Value: 1}}}, Property: &ast.StringNode{Value: "foo"}},
		},
		{
			"{foo:1}.bar",
			&ast.MemberNode{Node: &ast.MapNode{Pairs: []ast.Node{&ast.PairNode{Key: &ast.StringNode{Value: "foo"}, Value: &ast.IntegerNode{Value: 1}}}}, Property: &ast.StringNode{Value: "bar"}},
		},
		{
			"len(foo)",
			&ast.BuiltinNode{Name: "len", Arguments: []ast.Node{&ast.IdentifierNode{Value: "foo"}}},
		},
		{
			`foo matches "foo"`,
			&ast.MatchesNode{Left: &ast.IdentifierNode{Value: "foo"}, Right: &ast.StringNode{Value: "foo"}},
		},
		{
			`foo matches regex`,
			&ast.MatchesNode{Left: &ast.IdentifierNode{Value: "foo"}, Right: &ast.IdentifierNode{Value: "regex"}},
		},
		{
			`foo contains "foo"`,
			&ast.BinaryNode{Operator: "contains", Left: &ast.IdentifierNode{Value: "foo"}, Right: &ast.StringNode{Value: "foo"}},
		},
		{
			`foo startsWith "foo"`,
			&ast.BinaryNode{Operator: "startsWith", Left: &ast.IdentifierNode{Value: "foo"}, Right: &ast.StringNode{Value: "foo"}},
		},
		{
			`foo endsWith "foo"`,
			&ast.BinaryNode{Operator: "endsWith", Left: &ast.IdentifierNode{Value: "foo"}, Right: &ast.StringNode{Value: "foo"}},
		},
		{
			"1..9",
			&ast.BinaryNode{Operator: "..", Left: &ast.IntegerNode{Value: 1}, Right: &ast.IntegerNode{Value: 9}},
		},
		{
			"0 in []",
			&ast.BinaryNode{Operator: "in", Left: &ast.IntegerNode{}, Right: &ast.ArrayNode{Nodes: []ast.Node{}}},
		},
		{
			"not in_var",
			&ast.UnaryNode{Operator: "not", Node: &ast.IdentifierNode{Value: "in_var"}},
		},
		{
			"all(Tickets, {.Price > 0})",
			&ast.BuiltinNode{
				Name: "all", Arguments: []ast.Node{
					&ast.IdentifierNode{Value: "Tickets"}, &ast.ClosureNode{
						Node: &ast.BinaryNode{
							Operator: ">",
							Left:     &ast.MemberNode{Node: &ast.PointerNode{}, Property: &ast.StringNode{Value: "Price"}},
							Right:    &ast.IntegerNode{Value: 0},
						},
					},
				},
			},
		},
		{
			"one(Tickets, {#.Price > 0})",
			&ast.BuiltinNode{
				Name: "one",
				Arguments: []ast.Node{
					&ast.IdentifierNode{Value: "Tickets"},
					&ast.ClosureNode{
						Node: &ast.BinaryNode{
							Operator: ">",
							Left: &ast.MemberNode{
								Node:     &ast.PointerNode{},
								Property: &ast.StringNode{Value: "Price"},
							},
							Right: &ast.IntegerNode{Value: 0},
						},
					},
				},
			},
		},
		{
			"filter(Prices, {# > 100})",
			&ast.BuiltinNode{Name: "filter", Arguments: []ast.Node{&ast.IdentifierNode{Value: "Prices"}, &ast.ClosureNode{Node: &ast.BinaryNode{Operator: ">", Left: &ast.PointerNode{}, Right: &ast.IntegerNode{Value: 100}}}}},
		},
		{
			"array[1:2]",
			&ast.SliceNode{Node: &ast.IdentifierNode{Value: "array"}, From: &ast.IntegerNode{Value: 1}, To: &ast.IntegerNode{Value: 2}},
		},
		{
			"array[:2]",
			&ast.SliceNode{Node: &ast.IdentifierNode{Value: "array"}, To: &ast.IntegerNode{Value: 2}},
		},
		{
			"array[1:]",
			&ast.SliceNode{Node: &ast.IdentifierNode{Value: "array"}, From: &ast.IntegerNode{Value: 1}},
		},
		{
			"array[:]",
			&ast.SliceNode{Node: &ast.IdentifierNode{Value: "array"}},
		},
		{
			"[]",
			&ast.ArrayNode{},
		},
	}
	for _, test := range parseTests {
		is := is.New(t)
		actual, err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("%s:\n%v", test.input, err)
			continue
		}
		if m, ok := (actual.Node).(*ast.MatchesNode); ok {
			m.Regexp = nil
			actual.Node = m
		}
		is.Msg(test.input).Equal(ast.Dump(test.expected), ast.Dump(actual.Node))
	}
}

const errorTests = `
foo.
unexpected end of expression (1:4)
 | foo.
 | ...^

a+
unexpected token EOF (1:2)
 | a+
 | .^

a ? (1+2) c
unexpected token Identifier("c") (1:11)
 | a ? (1+2) c
 | ..........^

[a b]
unexpected token Identifier("b") (1:4)
 | [a b]
 | ...^

foo.bar(a b)
unexpected token Identifier("b") (1:11)
 | foo.bar(a b)
 | ..........^

{-}
a map key must be a quoted string, a number, a identifier, or an expression enclosed in parentheses (unexpected token Operator("-")) (1:2)
 | {-}
 | .^

a matches 'a:)b'
error parsing regexp: unexpected ): ` + "`a:)b`" + ` (1:16)
 | a matches 'a:)b'
 | ...............^

foo({.bar})
a map key must be a quoted string, a number, a identifier, or an expression enclosed in parentheses (unexpected token Operator(".")) (1:6)
 | foo({.bar})
 | .....^

.foo
cannot use pointer accessor outside closure (1:1)
 | .foo
 | ^

[1, 2, 3,,]
unexpected token Operator(",") (1:10)
 | [1, 2, 3,,]
 | .........^

[,]
unexpected token Operator(",") (1:2)
 | [,]
 | .^

{,}
a map key must be a quoted string, a number, a identifier, or an expression enclosed in parentheses (unexpected token Operator(",")) (1:2)
 | {,}
 | .^

{foo:1, bar:2, ,}
unexpected token Operator(",") (1:16)
 | {foo:1, bar:2, ,}
 | ...............^
`

func TestParse_error(t *testing.T) {
	tests := strings.Split(strings.Trim(errorTests, "\n"), "\n\n")
	for _, test := range tests {
		is := is.New(t)
		input := strings.SplitN(test, "\n", 2)
		if len(input) != 2 {
			t.Errorf("syntax error in test: %q", test)
			break
		}
		_, err := parser.Parse(input[0])
		if err == nil {
			err = fmt.Errorf("<nil>")
		}
		is.Msg(input[0]).Equal(input[1], err.Error())
	}
}

func TestParse_optional_chaining(t *testing.T) {
	parseTests := []struct {
		input    string
		expected ast.Node
	}{
		{
			"foo?.bar.baz",
			&ast.ChainNode{
				Node: &ast.MemberNode{
					Node: &ast.MemberNode{
						Node:     &ast.IdentifierNode{Value: "foo"},
						Property: &ast.StringNode{Value: "bar"},
						Optional: true,
					},
					Property: &ast.StringNode{Value: "baz"},
				},
			},
		},
		{
			"foo.bar?.baz",
			&ast.ChainNode{
				Node: &ast.MemberNode{
					Node: &ast.MemberNode{
						Node:     &ast.IdentifierNode{Value: "foo"},
						Property: &ast.StringNode{Value: "bar"},
					},
					Property: &ast.StringNode{Value: "baz"},
					Optional: true,
				},
			},
		},
		{
			"foo?.bar?.baz",
			&ast.ChainNode{
				Node: &ast.MemberNode{
					Node: &ast.MemberNode{
						Node:     &ast.IdentifierNode{Value: "foo"},
						Property: &ast.StringNode{Value: "bar"},
						Optional: true,
					},
					Property: &ast.StringNode{Value: "baz"},
					Optional: true,
				},
			},
		},
		{
			"!foo?.bar.baz",
			&ast.UnaryNode{
				Operator: "!",
				Node: &ast.ChainNode{
					Node: &ast.MemberNode{
						Node: &ast.MemberNode{
							Node:     &ast.IdentifierNode{Value: "foo"},
							Property: &ast.StringNode{Value: "bar"},
							Optional: true,
						},
						Property: &ast.StringNode{Value: "baz"},
					},
				},
			},
		},
		{
			"foo.bar[a?.b]?.baz",
			&ast.ChainNode{
				Node: &ast.MemberNode{
					Node: &ast.MemberNode{
						Node: &ast.MemberNode{
							Node:     &ast.IdentifierNode{Value: "foo"},
							Property: &ast.StringNode{Value: "bar"},
						},
						Property: &ast.ChainNode{
							Node: &ast.MemberNode{
								Node:     &ast.IdentifierNode{Value: "a"},
								Property: &ast.StringNode{Value: "b"},
								Optional: true,
							},
						},
					},
					Property: &ast.StringNode{Value: "baz"},
					Optional: true,
				},
			},
		},
	}
	for _, test := range parseTests {
		is := is.New(t)
		actual, err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("%s:\n%v", test.input, err)
			continue
		}
		if m, ok := (actual.Node).(*ast.MatchesNode); ok {
			m.Regexp = nil
			actual.Node = m
		}
		is.Msg(test.input).Equal(ast.Dump(test.expected), ast.Dump(actual.Node))
	}
}
