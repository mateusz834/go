package ssaparser

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"text/scanner"
)

// TODO: make sure that the order of Preds in blocks is the same as would be created by AddEdge.

func ParseSSAFunc(s string) (*Func, error) {
	p := &parser{
		lines: slices.Collect(strings.Lines(s)),
		tok:   eol,
	}

	// skip empty lines
	for p.tok == eol {
		p.next()
	}

	blocks, err := p.blocks()
	if err != nil {
		return nil, err
	}

	return &Func{
		Blocks: blocks,
	}, nil
}

type parser struct {
	lines []string

	s      scanner.Scanner
	lineNo int

	tok rune
	lit string
	pos scanner.Position

	prevEOF bool
}

const eol rune = math.MinInt32

func tokenString(tok rune) string {
	if tok == eol {
		return "EOL"
	}
	return scanner.TokenString(tok)
}

func (p *parser) nextLine() {
	var line string
	if len(p.lines) != 0 {
		line = p.lines[0]
		p.lines = p.lines[1:]
	}
	p.s.Init(strings.NewReader(line))
	p.s.Mode = scanner.GoTokens &^ scanner.SkipComments
	p.s.Filename = "ssa"
	p.lineNo++
}

func (p *parser) next() {
	if p.tok == eol {
		p.nextLine()
	}
	p.tok = p.s.Scan()
	p.lit = p.s.TokenText()
	p.pos = p.s.Position
	p.pos.Line += p.lineNo
	if p.tok == scanner.EOF {
		p.tok = eol
	}
	//fmt.Printf("%v %q %v\n", tokenString(p.tok), p.lit, p.pos)
}

func (p *parser) blocks() ([]*Block, error) {
	var blocks []*Block
	for p.tok == scanner.Ident {
		fmt.Printf("p.tok: %v\n", p.tok)
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
		//fmt.Printf("%#v %v", b, err)
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
	lit := p.lit
	p.next()
	if err := p.expectTok(r); err != nil {
		return "", err
	}
	return lit, nil
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
