package optimizer

import (
	. "github.com/ilius/expr/ast"
	"github.com/ilius/expr/conf"
)

func Optimize(node *Node, config *conf.Config) error {
	Walk(node, &inArray{})
	for limit := 1000; limit >= 0; limit-- {
		fold := &fold{}
		Walk(node, fold)
		if fold.err != nil {
			return fold.err
		}
		if !fold.applied {
			break
		}
	}
	if config != nil && len(config.ConstFns) > 0 {
		for limit := 100; limit >= 0; limit-- {
			constExpr := &constExpr{
				fns: config.ConstFns,
			}
			Walk(node, constExpr)
			if constExpr.err != nil {
				return constExpr.err
			}
			if !constExpr.applied {
				break
			}
		}
	}
	Walk(node, &inRange{})
	Walk(node, &constRange{})
	return nil
}
