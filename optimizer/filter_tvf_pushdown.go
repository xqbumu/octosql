package optimizer

import (
	"github.com/cube2222/octosql/octosql"
	. "github.com/cube2222/octosql/physical"
)

func PushDownFilterPredicatesToTvf(node Node) (Node, bool) {
	changed := false
	t := Transformers{
		NodeTransformer: func(node Node) Node {
			if node.NodeType != NodeTypeFilter {
				return node
			}
			if node.Filter.Source.NodeType != NodeTypeTableValuedFunction {
				return node
			}
			source, ok := node.Filter.Source.TableValuedFunction.Arguments["source"]
			if !ok {
				return node
			}
			if source.Table.Table.NodeType != NodeTypeDatasource {
				return node
			}

			filterPredicates := node.Filter.Predicate.SplitByAnd()
			alreadyPushedDown := source.Table.Table.Datasource.Predicates

			newFilterPredicates, newPushedDownPredicates, curChanged := source.Table.Table.Datasource.PushDownPredicates(filterPredicates, alreadyPushedDown)
			if !curChanged {
				return node
			}
			changed = true

			out := Node{
				Schema:   node.Filter.Source.Schema,
				NodeType: NodeTypeDatasource,
				Datasource: &Datasource{
					Name:                     source.Table.Table.Datasource.Name,
					Alias:                    source.Table.Table.Datasource.Alias,
					DatasourceImplementation: source.Table.Table.Datasource.DatasourceImplementation,
					Predicates:               newPushedDownPredicates,
					VariableMapping:          source.Table.Table.Datasource.VariableMapping,
				},
			}
			source.Table.Table = out
			node.Filter.Source.TableValuedFunction.Arguments["source"] = source

			if len(newFilterPredicates) > 0 {
				out = Node{
					Schema:   node.Schema,
					NodeType: NodeTypeFilter,
					Filter: &Filter{
						Predicate: Expression{
							Type:           octosql.Boolean,
							ExpressionType: ExpressionTypeAnd,
							And: &And{
								Arguments: newFilterPredicates,
							},
						},
						Source: node,
					},
				}
			}

			return node
		},
	}
	output := t.TransformNode(node)

	if changed {
		return output, true
	} else {
		return node, false
	}
}
