package logical

import (
	"context"

	"github.com/cube2222/octosql/physical"
)

type With struct {
	cteNames []string
	cteNodes []Node
	source   Node
}

func NewWith(cteNames []string, cteNodes []Node, source Node) *With {
	return &With{
		cteNames: cteNames,
		cteNodes: cteNodes,
		source:   source,
	}
}

func (node *With) Typecheck(ctx context.Context, env physical.Environment, logicalEnv Environment) (physical.Node, map[string]string) {
	newCTEs := make(map[string]CommonTableExpr)
	for k, v := range logicalEnv.CommonTableExprs {
		newCTEs[k] = v
	}

	for i := range node.cteNodes {
		cte, mapping := node.cteNodes[i].Typecheck(ctx, env, Environment{
			CommonTableExprs:     newCTEs,
			TableValuedFunctions: logicalEnv.TableValuedFunctions,
			UniqueVariableNames:  logicalEnv.UniqueVariableNames,
			UniqueNameGenerator:  logicalEnv.UniqueNameGenerator,
		})
		newCTEs[node.cteNames[i]] = CommonTableExpr{
			Node:                  cte,
			UniqueVariableMapping: mapping,
		}
	}

	return node.source.Typecheck(ctx, env, Environment{
		CommonTableExprs:     newCTEs,
		TableValuedFunctions: logicalEnv.TableValuedFunctions,
		UniqueVariableNames:  logicalEnv.UniqueVariableNames,
		UniqueNameGenerator:  logicalEnv.UniqueNameGenerator,
	})
}
