package ssaparser

import "strings"

func PrintFunc(f *Func) string {
	var s strings.Builder
	for i, b := range f.Blocks {
		if i != 0 {
			s.WriteString("\n")
		}
		writeBlock(&s, b)
		s.WriteString("\n")
	}
	return s.String()
}

func PrintBlock(b *Block) string {
	var s strings.Builder
	writeBlock(&s, b)
	return s.String()
}

func PrintValue(v *Value) string {
	var s strings.Builder
	writeValue(&s, v)
	return s.String()
}

func writeBlock(s *strings.Builder, b *Block) {
	s.WriteString(b.Name)
	s.WriteString(":")

	if len(b.Preds) != 0 {
		s.WriteString(" <-")
		for _, v := range b.Preds {
			s.WriteString(" ")
			s.WriteString(v)
		}
	}

	if b.Entry {
		s.WriteString(" (entry)")
	}

	for _, v := range b.Values {
		s.WriteString("\n    ")
		writeValue(s, v)
	}

	s.WriteString("\n")
	s.WriteString(b.Kind)
	for _, c := range b.Controls {
		s.WriteString(" ")
		s.WriteString(c)
	}
	if len(b.Succ) != 0 {
		s.WriteString(" ->")
		for _, v := range b.Succ {
			s.WriteString(" ")
			s.WriteString(v)
		}
	}
}

func writeValue(s *strings.Builder, v *Value) {
	s.WriteString("(")
	s.WriteString(v.Pos)
	s.WriteString(") ")

	s.WriteString(v.Name)
	s.WriteString(" = ")
	s.WriteString(v.Op)

	s.WriteString(" <")
	s.WriteString(v.Type)
	s.WriteString(">")

	if v.AuxInt != "" {
		s.WriteString(" [")
		s.WriteString(v.AuxInt)
		s.WriteString("]")
	}

	if v.Aux != "" {
		s.WriteString(" {")
		s.WriteString(v.Aux)
		s.WriteString("}")
	}

	for _, a := range v.Args {
		s.WriteString(" ")
		s.WriteString(a)
	}
}
