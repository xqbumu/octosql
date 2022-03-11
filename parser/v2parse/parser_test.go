package v2parse

import (
	"log"
	"testing"

	"github.com/cube2222/octosql/parser"
	v2 "github.com/cube2222/octosql/parser/v2"
)

func TestParseNode(t *testing.T) {
	sql := "SELECT * FROM t"
	statement, err := parser.Parse(sql)
	if err != nil {
		t.Errorf("couldn't parse query: %w", err)
	}
	logicalPlan, outputOptions, err := ParseNode(statement.(v2.SelectStatement), true)
	if err != nil {
		t.Errorf("couldn't parse query: %w", err)
	}

	log.Println(logicalPlan, outputOptions)
}
