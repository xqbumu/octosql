package v2

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggerExprs(t *testing.T) {
	tree, err := Parse("SELECT window_end, user_id, COUNT(*) as clicks FROM with_tumble GROUP BY window_end, user_id TRIGGER ON WATERMARK, COUNTING 500")
	require.NoError(t, err)

	treeSel, ok := tree.(*Select)
	require.True(t, ok)

	exprs := treeSel.TriggerExprs
	sel := &Select{}
	for i := 0; i < len(exprs); i++ {
		sel.AddTriggerExpr(exprs[i])
	}
	buf := NewTrackedBuffer(nil)
	sel.TriggerExprs.Format(buf)
	log.Println(buf.String())
	want := "TRIGGER ON WATERMARK, COUNTING 500"
	if buf.String() != want {
		t.Errorf("where: %q, want %s", buf.String(), want)
	}
}
