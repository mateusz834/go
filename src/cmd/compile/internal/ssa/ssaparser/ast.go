package ssaparser

type Func struct {
	Blocks []*Block
}

type Block struct {
	Entry bool

	Name  string
	Preds []string

	Values []*Value

	Kind     string
	Controls []string
	Succ     []string
}

type Value struct {
	Pos    string
	Name   string
	Op     string
	Type   string
	Aux    string
	AuxInt string
	Args   []string
}
