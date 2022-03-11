package v2

// TableValuedFunctionArguments represents SELECT expressions.
type TableValuedFunctionArguments []*TableValuedFunctionArgument

// Format formats the node.
func (node TableValuedFunctionArguments) Format(buf *TrackedBuffer) {
	var prefix string
	for _, n := range node {
		buf.astPrintf(node, "%s%v", prefix, n)
		prefix = ", "
	}
}

// Format formats the node.
func (node TableValuedFunctionArguments) formatFast(buf *TrackedBuffer) {
	var prefix string
	for _, n := range node {
		buf.astPrintf(node, "%s%v", prefix, n)
		prefix = ", "
	}
}

// TableValuedFunctionArgument defines an aliased SELECT expression.
type TableValuedFunctionArgument struct {
	Name  ColIdent
	Value TableValuedFunctionArgumentValue
}

// Format formats the node.
func (node *TableValuedFunctionArgument) Format(buf *TrackedBuffer) {
	buf.astPrintf(node, "%v => %v", node.Name, node.Value)
}

// formatFast formats the node.
func (node *TableValuedFunctionArgument) formatFast(buf *TrackedBuffer) {
	node.Name.formatFast(buf)
	buf.WriteString(" => ")
	node.Value.formatFast(buf)
}

// TableValuedFunctionArgumentValue defines table valued function argument value.
type TableValuedFunctionArgumentValue interface {
	iTableValuedFunctionArgumentValue()
	SQLNode
}

func (*ExprTableValuedFunctionArgumentValue) iTableValuedFunctionArgumentValue()            {}
func (*TableDescriptorTableValuedFunctionArgumentValue) iTableValuedFunctionArgumentValue() {}
func (*FieldDescriptorTableValuedFunctionArgumentValue) iTableValuedFunctionArgumentValue() {}

type TableValuedFunction struct {
	Name ColIdent
	Args TableValuedFunctionArguments
	As   TableIdent
}

func (*TableValuedFunction) iTableExpr() {}

// Format formats the node.
func (node *TableValuedFunction) Format(buf *TrackedBuffer) {
	buf.astPrintf(node, "%v(%v) as %v", node.Name, node.Args, node.As)
}

// formatFast formats the node.
func (node *TableValuedFunction) formatFast(buf *TrackedBuffer) {
	node.Name.formatFast(buf)
	buf.WriteByte('(')
	node.Args.formatFast(buf)
	buf.WriteByte(')')
	if !node.As.IsEmpty() {
		buf.WriteString(" as ")
		node.As.formatFast(buf)
	}
}

type ExprTableValuedFunctionArgumentValue struct {
	Expr Expr
}

func (node *ExprTableValuedFunctionArgumentValue) Format(buf *TrackedBuffer) {
	buf.astPrintf(node, "%v", node.Expr)
}

// formatFast formats the node.
func (node *ExprTableValuedFunctionArgumentValue) formatFast(buf *TrackedBuffer) {
	node.Expr.formatFast(buf)
}

type FieldDescriptorTableValuedFunctionArgumentValue struct {
	Field *ColName
}

func (node *FieldDescriptorTableValuedFunctionArgumentValue) Format(buf *TrackedBuffer) {
	buf.astPrintf(node, "DESCRIPTOR(%v)", node.Field)
}

// formatFast formats the node.
func (node *FieldDescriptorTableValuedFunctionArgumentValue) formatFast(buf *TrackedBuffer) {
	buf.WriteString("DESCRIPTOR")
	buf.WriteByte('(')
	node.Field.formatFast(buf)
	buf.WriteByte(')')
}

type TableDescriptorTableValuedFunctionArgumentValue struct {
	Table TableExpr
}

func (node *TableDescriptorTableValuedFunctionArgumentValue) Format(buf *TrackedBuffer) {
	buf.astPrintf(node, "TABLE(%v)", node.Table)
}

// formatFast formats the node.
func (node *TableDescriptorTableValuedFunctionArgumentValue) formatFast(buf *TrackedBuffer) {
	buf.WriteString("TABLE")
	buf.WriteByte('(')
	node.Table.formatFast(buf)
	buf.WriteByte(')')
}
