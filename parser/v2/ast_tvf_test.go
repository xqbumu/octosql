package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTvf(t *testing.T) {
	// yyDebug = 4
	tree, err := Parse("SELECT * FROM tvf_range(start=>1,end=>10) r")
	require.NoError(t, err)

	treeSel, ok := tree.(*Select)
	require.True(t, ok)

	buf := NewTrackedBuffer(nil)
	treeSel.Format(buf)
	want := "select * from tvf_range(`start` => 1, `end` => 10) as r"
	if buf.String() != want {
		t.Errorf("where: %q, want %s", buf.String(), want)
	}
}
