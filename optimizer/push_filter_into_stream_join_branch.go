package optimizer

import (
	"github.com/cube2222/octosql/octosql"
	. "github.com/cube2222/octosql/physical"
)

func PushDownFilterPredicatesIntoStreamJoinBranch(node Node) (Node, bool) {
	changed := false
	t := Transformers{
		NodeTransformer: func(node Node) Node {
			if node.NodeType != NodeTypeFilter {
				return node
			}
			if node.Filter.Source.NodeType != NodeTypeStreamJoin {
				return node
			}

			// Directly PushDown
			leftSchema := node.Filter.Source.StreamJoin.Left.Schema
			rightSchema := node.Filter.Source.StreamJoin.Right.Schema

			filterPredicates := node.Filter.Predicate.SplitByAnd()
			var stayedAbove, pushedDownLeft, pushedDownRight []Expression

			for i := range filterPredicates {
				variablesUsed := filterPredicates[i].VariablesUsed()
				// Clarification:
				// If it doesn't use variables from any join branch,
				// then it gets pushed down into both.
				usesLeftBranch := usesVariablesFromSchema(leftSchema, variablesUsed)
				usesRightBranch := usesVariablesFromSchema(rightSchema, variablesUsed)
				if !usesLeftBranch {
					pushedDownRight = append(pushedDownRight, filterPredicates[i])
				}
				if !usesRightBranch {
					pushedDownLeft = append(pushedDownLeft, filterPredicates[i])
				}
				if usesLeftBranch && usesRightBranch {
					stayedAbove = append(stayedAbove, filterPredicates[i])
				}
			}

			// Cross PushDown
			leftJoinKeys := node.Filter.Source.StreamJoin.LeftKey
			rightJoinKeys := node.Filter.Source.StreamJoin.RightKey
			genAppender := func(variable *Variable) (func(expr Expression), *Variable) {
				for i := 0; i < len(leftJoinKeys); i++ {
					if compareVariable(leftJoinKeys[i].Variable, variable) {
						appender := func(expr Expression) {
							pushedDownRight = append(pushedDownRight, expr)
						}
						return appender, rightJoinKeys[i].Variable
					}
					if compareVariable(rightJoinKeys[i].Variable, variable) {
						appender := func(expr Expression) {
							pushedDownLeft = append(pushedDownLeft, expr)
						}
						return appender, leftJoinKeys[i].Variable
					}
				}
				return nil, nil
			}
			for i := range filterPredicates {
				expr := filterPredicates[i]
				switch expr.ExpressionType {
				case ExpressionTypeFunctionCall:
					if appender, variable := genAppender(expr.FunctionCall.Arguments[0].Variable); variable != nil {
						expr := Expression{
							Type:           expr.Type,
							ExpressionType: expr.ExpressionType,
							FunctionCall: func() *FunctionCall {
								fc := *expr.FunctionCall
								fc.Arguments = append([]Expression{}, fc.Arguments...)
								fc.Arguments[0].Variable = variable
								return &fc
							}(),
						}
						appender(expr)
					}
				}
			}

			if len(stayedAbove) == len(filterPredicates) {
				return node
			}
			changed = true

			joinSourceLeft := node.Filter.Source.StreamJoin.Left
			if len(pushedDownLeft) > 0 {
				joinSourceLeft = Node{
					Schema:   joinSourceLeft.Schema,
					NodeType: NodeTypeFilter,
					Filter: &Filter{
						Source: joinSourceLeft,
						Predicate: Expression{
							Type:           octosql.Boolean,
							ExpressionType: ExpressionTypeAnd,
							And: &And{
								Arguments: pushedDownLeft,
							},
						},
					},
				}
			}
			joinSourceRight := node.Filter.Source.StreamJoin.Right
			if len(pushedDownRight) > 0 {
				joinSourceRight = Node{
					Schema:   joinSourceRight.Schema,
					NodeType: NodeTypeFilter,
					Filter: &Filter{
						Source: joinSourceRight,
						Predicate: Expression{
							Type:           octosql.Boolean,
							ExpressionType: ExpressionTypeAnd,
							And: &And{
								Arguments: pushedDownRight,
							},
						},
					},
				}
			}

			out := Node{
				Schema:   node.Filter.Source.Schema,
				NodeType: NodeTypeStreamJoin,
				StreamJoin: &StreamJoin{
					LeftKey:  node.Filter.Source.StreamJoin.LeftKey,
					RightKey: node.Filter.Source.StreamJoin.RightKey,
					Left:     joinSourceLeft,
					Right:    joinSourceRight,
				},
			}
			if len(stayedAbove) > 0 {
				out = Node{
					Schema:   out.Schema,
					NodeType: NodeTypeFilter,
					Filter: &Filter{
						Predicate: Expression{
							Type:           octosql.Boolean,
							ExpressionType: ExpressionTypeAnd,
							And: &And{
								Arguments: stayedAbove,
							},
						},
						Source: out,
					},
				}
			}

			return out
		},
	}
	output := t.TransformNode(node)

	if changed {
		return output, true
	} else {
		return node, false
	}
}

func usesVariablesFromSchema(schema Schema, variables []string) bool {
	for _, name := range variables {
		for _, field := range schema.Fields {
			if VariableNameMatchesField(name, field.Name) {
				return true
			}
		}
	}
	return false
}

func compareVariable(left, right *Variable) bool {
	return left.Name == right.Name && left.IsLevel0 == right.IsLevel0
}
