package ssaparser

import (
	"fmt"
	"math"
	"strings"
	"text/scanner"
)

// TODO: make sure that the order of Preds in blocks is the same as would be created by AddEdge.

func ParseSSAFunc(s string) (*Func, error) {
	var p parser
	// TODO: error handle
	p.s.Init(strings.NewReader(s))
	p.src = s
	p.s.Mode = scanner.GoTokens &^ scanner.SkipComments
	p.s.Filename = "ssa"
	p.pos.Line = 1
	p.next()

	blocks, err := p.blocks()
	if err != nil {
		return nil, err
	}

	return &Func{
		Blocks: blocks,
	}, nil
}

type parser struct {
	src string
	s   scanner.Scanner

	tok rune
	lit string
	pos scanner.Position

	prevTok rune
	prevEOL bool

	prevEOF bool
}

const eol rune = math.MinInt32

func tokenString(tok rune) string {
	if tok == eol {
		return "EOL"
	}
	return scanner.TokenString(tok)
}

func (p *parser) next0() {
	//defer func() {
	//	fmt.Printf("next(): %v %q %v\n", tokenString(p.tok), p.lit, p.pos)
	//}()

	prevTokLine := p.pos.Line
	tok := p.prevTok
	if !p.prevEOL {
		tok = p.s.Scan()
	}

	if prevTokLine == p.s.Line {
		p.tok = tok
		p.lit = p.s.TokenText()
		p.pos = p.s.Position
		p.prevEOL = false
		return
	}

	p.prevTok = tok
	p.prevEOL = true
	p.tok = eol
	p.lit = ""
	p.pos = p.s.Position
}

// Allow "// comments" directly before EOL or after EOL.
func (p *parser) next() {
	p.next0()
	//for {
	//	p.next0()
	//	if p.tok == scanner.Comment {
	//		continue
	//	}
	//}
}

func (p *parser) blocks() ([]*Block, error) {
	var blocks []*Block
	for p.tok == scanner.Ident {
		b, err := p.block()
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (p *parser) block() (*Block, error) {
	blockName := p.lit
	blockPos := p.pos

	if err := p.expectTok(scanner.Ident); err != nil {
		return nil, err
	}
	if err := p.expectTok(':'); err != nil {
		return nil, err
	}

	entry := false
	if p.tok == '(' {
		m, err := p.between('(', ')')
		if err != nil {
			return nil, err
		}
		if m == "entry" {
			entry = true
		} else {
			return nil, fmt.Errorf("unexpected value between (): %q", m)
		}
	}

	var preds []string
	if p.tok == '<' {
		p.next()
		if err := p.expectTok('-'); err != nil {
			return nil, err
		}
		preds = p.idents()
	}

	if err := p.expectTok(eol); err != nil {
		return nil, err
	}

	vals, err := p.values()
	if err != nil {
		return nil, fmt.Errorf("block %v (%v): %v", blockName, blockPos, err)
	}

	// TODO: aux and stuff
	exitKind := p.lit

	p.next()
	controls := p.idents()

	var succs []string
	if p.tok == '-' {
		p.next()
		if err := p.expectTok('>'); err != nil {
			return nil, err
		}
		succs = p.idents()
	}

	if err := p.expectTok(eol); err != nil {
		return nil, err
	}

	return &Block{
		Entry:    entry,
		Name:     blockName,
		Preds:    preds,
		Values:   vals,
		Kind:     exitKind,
		Controls: controls,
		Succ:     succs,
	}, nil
}

func (p *parser) values() ([]*Value, error) {
	var values []*Value
	for p.tok == '(' {
		b, err := p.value()
		if err != nil {
			return nil, fmt.Errorf("value: %v", err)
		}
		values = append(values, b)
	}
	return values, nil
}

func (p *parser) value() (*Value, error) {
	pos, err := p.between('(', ')')
	if err != nil {
		return nil, err
	}

	name := p.lit
	p.next()

	if err := p.expectTok('='); err != nil {
		return nil, err
	}

	op := p.lit
	p.next()

	typ, err := p.between('<', '>')
	if err != nil {
		return nil, fmt.Errorf("parsing type: %v", err)
	}

	var aux string
	var auxInt string
outer:
	for {
		switch p.tok {
		case '[':
			auxInt, err = p.between('[', ']')
			if err != nil {
				return nil, fmt.Errorf("parsing aux: %v", err)
			}
		case '{':
			aux, err = p.between('{', '}')
			if err != nil {
				return nil, fmt.Errorf("parsing aux: %v", err)
			}
		default:
			break outer
		}
	}

	args := p.idents()

	names := ""
	if p.tok == '(' {
		names, err = p.between('(', ')')
		if err != nil {
			return nil, fmt.Errorf("parsing aux: %v", err)
		}
	}
	_ = names

	if err := p.expectTok(eol); err != nil {
		return nil, err
	}

	return &Value{
		Pos:    pos,
		Name:   name,
		Op:     op,
		Type:   typ,
		Aux:    aux,
		AuxInt: auxInt,
		Args:   args,
	}, nil
}

func (p *parser) idents() []string {
	succs := []string{}
	for p.tok == scanner.Ident {
		succs = append(succs, p.lit)
		p.next()
	}
	return succs
}

func (p *parser) between(l, r rune) (string, error) {
	if err := p.expectTok(l); err != nil {
		return "", err
	}

	startOff := p.pos.Offset
	for p.tok != r && p.tok != eol {
		p.next()
	}
	endOff := p.pos.Offset

	if err := p.expectTok(r); err != nil {
		return "", err
	}

	return p.src[startOff:endOff], nil
}

func (p *parser) expect(got, want rune) error {
	return fmt.Errorf("%v: got %v; want %v", p.pos, tokenString(got), tokenString(want))
}

func (p *parser) expectTok(want rune) error {
	if p.tok != want {
		return p.expect(p.tok, want)
	}
	p.next()
	return nil
}
