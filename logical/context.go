package logical

type contextKey int

const (
	ctxKeyUndefined contextKey = iota
	ctxKeyFromWith
)
