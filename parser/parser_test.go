package parser

import (
	"log"
	"testing"

	v1 "github.com/cube2222/octosql/parser/v1"
	v1parse "github.com/cube2222/octosql/parser/v1parse"
	v2 "github.com/cube2222/octosql/parser/v2"
	v2parse "github.com/cube2222/octosql/parser/v2parse"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	sql := "select a2 from ((t1 join t2) join t3 on b=c1) join t4"

	v1Statement, err := v1.Parse(sql)
	if err != nil {
		t.Errorf("couldn't parse query: %s", err)
	}

	v2Statement, err := v2.Parse(sql)
	if err != nil {
		t.Errorf("couldn't parse query: %s", err)
	}

	v1Buf := v1.NewTrackedBuffer(nil)
	v2Buf := v2.NewTrackedBuffer(nil)

	v1Statement.Format(v1Buf)
	v2Statement.Format(v2Buf)

	assert.Equal(t, v1Buf.String(), v2Buf.String())

	log.Println(v1Buf.String())
	log.Println(v2Buf.String())

	v1LP, v1OO, v1Err := v1parse.ParseNode(v1Statement.(v1.SelectStatement), true)
	if v1Err != nil {
		t.Errorf("couldn't parse query: %s", v1Err)
	}
	log.Println(v1LP, v1OO)

	v2LP, v2OO, v2Err := v2parse.ParseNode(v2Statement.(v2.SelectStatement), true)
	if v2Err != nil {
		t.Errorf("couldn't parse query: %s", v2Err)
	}
	log.Println(v2LP, v2OO)
}

func TestComment(t *testing.T) {
	v2.Debug(4)
	sql := "select /*+ AA(a) */ 1"

	statement, err := v2.Parse(sql)
	if err != nil {
		t.Errorf("couldn't parse query: %s", err)
	}

	log.Println(statement)
}

func compareSelect(a v1.SelectStatement, b v2.SelectStatement) bool {
	switch aTyped := a.(type) {
	case *v1.Select:
		if bTyped, ok := b.(*v2.Select); ok {
			return compareSelectExprs(aTyped.SelectExprs, bTyped.SelectExprs)
		}
		return false

	// case *v1.Union:
	// 	plan, err := ParseUnion(statement)
	// 	return plan, &logical.OutputOptions{}, err

	// [TODO]
	// case *v1.ParenSelect:
	// 	return ParseNode(statement.Select, topmost)

	// [TODO]
	// case *v1.With:
	// 	return ParseWith(statement, topmost)

	default:
		return false
	}
}

func compareSelectExprs(a v1.SelectExprs, b v2.SelectExprs) bool {
	if len(a) != len(b) {
		return false
	}

	aBuf := v1.NewTrackedBuffer(nil)
	bBuf := v2.NewTrackedBuffer(nil)

	for i := 0; i < len(a); i++ {
		aBuf.Reset()
		bBuf.Reset()
		a[i].Format(aBuf)
		b[i].Format(bBuf)
		if aBuf.String() != bBuf.String() {
			return false
		}
	}

	return false
}
