package parser

import (
	"github.com/cube2222/octosql/logical"
	v2 "github.com/cube2222/octosql/parser/v2"
	v2parse "github.com/cube2222/octosql/parser/v2parse"
)

var (
	Parse = v2.Parse
)

func LogicalPlan(statement v2.Statement, topmost bool) (logical.Node, *v2parse.OutputOptions, error) {
	return v2parse.ParseNode(statement.(v2.SelectStatement), true)
}
