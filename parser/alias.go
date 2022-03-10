package parser

import (
	"github.com/cube2222/octosql/logical"
	v1 "github.com/cube2222/octosql/parser/v1"
	v1parse "github.com/cube2222/octosql/parser/v1parse"
)

var (
	Parse = v1.Parse
)

func LogicalPlan(statement v1.Statement, topmost bool) (logical.Node, *v1parse.OutputOptions, error) {
	return v1parse.ParseNode(statement.(v1.SelectStatement), true)
}
