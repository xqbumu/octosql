package flags

import (
	"flag"
)

var (
	// TruncateUILen truncate queries in debug UIs to the given length. 0 means unlimited.
	TruncateUILen = flag.Int("sql-max-length-ui", 512, "truncate queries in debug UIs to the given length (default 512)")

	// TruncateErrLen truncate queries in error logs to the given length. 0 means unlimited.
	TruncateErrLen = flag.Int("sql-max-length-errors", 0, "truncate queries in error logs to the given length (default unlimited)")
)
