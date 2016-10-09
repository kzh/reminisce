package vm

type Node interface {
    Location() int
}

type node struct {
    line int
}

func (n *node) Location() int {
    return n.line
}

func (n *node) SetLocation(line int) {
    n.line = line
}

type SectionDataNode struct {
    node
    labels []ValuedLabel
}

type ValuedLabel struct {
    node
    Name        string
    Size, Value int
}

type SectionTextNode struct {
    node
    Instructions []*InstructionNode
    Labels       map[string]int
}

type InstructionNode struct {
    node
    Name      string
    Arguments []Node
}

type AccessNode struct {
    node
    Size       int
    Expression Node
}

type LabelNode struct {
    node
    Name string
}

type RegisterNode struct {
    node
    Register string
}

type BinopExprNode struct {
    node
    Operator    string
    Left, Right Node
}

type ConstantNode struct {
    node
    Value int64
}
