package v2

// TriggerExprs represents a TRIGGER clause.
type TriggerExprs []TriggerExpr

// Format formats the node.
func (node TriggerExprs) Format(buf *TrackedBuffer) {
	prefix := "TRIGGER "
	for _, n := range node {
		buf.Myprintf("%s%v", prefix, n)
		prefix = ", "
	}
}

type TriggerExpr interface {
	iTriggerExpr()
	SQLNode
}

func (node *WatermarkTriggerExpr) iTriggerExpr()   {}
func (node *EndOfStreamTriggerExpr) iTriggerExpr() {}
func (node *DelayTriggerExpr) iTriggerExpr()       {}
func (node *CountingTriggerExpr) iTriggerExpr()    {}

type WatermarkTriggerExpr struct {
}

func (node *WatermarkTriggerExpr) Format(buf *TrackedBuffer) {
	buf.Myprintf("ON WATERMARK")
}

// formatFast formats the node.
func (node *WatermarkTriggerExpr) formatFast(buf *TrackedBuffer) {
	buf.WriteString("ON WATERMARK")
}

type EndOfStreamTriggerExpr struct {
}

func (nodoe *EndOfStreamTriggerExpr) Format(buf *TrackedBuffer) {
	buf.Myprintf("ON END OF STREAM")
}

// formatFast formats the node.
func (node *EndOfStreamTriggerExpr) formatFast(buf *TrackedBuffer) {
	buf.WriteString("ON END OF STREAM")
}

type DelayTriggerExpr struct {
	Delay Expr
}

func (node *DelayTriggerExpr) Format(buf *TrackedBuffer) {
	buf.Myprintf("DELAY %v", node.Delay)
}

// formatFast formats the node.
func (node *DelayTriggerExpr) formatFast(buf *TrackedBuffer) {
	buf.WriteString("DELAY ")
	node.Delay.formatFast(buf)
}

type CountingTriggerExpr struct {
	Count Expr
}

func (node *CountingTriggerExpr) Format(buf *TrackedBuffer) {
	buf.Myprintf("COUNTING %v", node.Count)
}

// formatFast formats the node.
func (node *CountingTriggerExpr) formatFast(buf *TrackedBuffer) {
	buf.WriteString("COUNTING ")
	node.Count.formatFast(buf)
}

// EqualsTriggerExpr does deep equals between the two objects.
func EqualsTriggerExpr(inA, inB TriggerExpr) bool {
	if inA == nil && inB == nil {
		return true
	}
	if inA == nil || inB == nil {
		return false
	}
	switch a := inA.(type) {
	case *WatermarkTriggerExpr:
		_, ok := inB.(*WatermarkTriggerExpr)
		if !ok {
			return false
		}
		return true
	case *EndOfStreamTriggerExpr:
		_, ok := inB.(*EndOfStreamTriggerExpr)
		if !ok {
			return false
		}
		return true
	case *DelayTriggerExpr:
		b, ok := inB.(*DelayTriggerExpr)
		if !ok {
			return false
		}
		return EqualsExpr(a.Delay, b.Delay)
	case *CountingTriggerExpr:
		b, ok := inB.(*CountingTriggerExpr)
		if !ok {
			return false
		}
		return EqualsExpr(a.Count, b.Count)
	default:
		// this should never happen
		return false
	}
}
