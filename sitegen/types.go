package sitegen

// SortKey represents a field and sort direction.
type SortKey struct {
	Path string
	Asc  bool
}

// Binding defines a variable binding from header.
type Binding struct {
	Name   string
	Target string
	URL    string
	Lazy   bool
	Glob   bool
	Manual bool
}

// Section is a simple markdown section.
type Section struct {
	ID   string
	Text string
}

// Loop represents a for-loop block.
type Loop struct {
	Vars   []string
	Source string
	Sort   []SortKey
	Body   []Node
}

// Node is a body node.
type Node interface{}

// Field line in loop body.
type Field struct {
	Path string
}
