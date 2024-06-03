package parser

import (
	"go/ast"
	"go/token"
)

func (p *parser) parseTgoStmt() (s ast.Stmt) {
	switch p.tok {
	case token.STRING_TEMPLATE:
		templLit := p.parseTemplateLiteral()
		p.expectSemi()
		return &ast.ExprStmt{X: templLit}
	case token.LSS:
		openPos := p.pos
		p.next()

		closing := false
		if p.tok == token.QUO {
			closing = true
			p.next()
		}

		ident := p.parseIdent()

		if closing {
			closePos := p.expect2(token.GTR)
			return &ast.EndTagStmt{
				OpenPos:  openPos,
				Name:     ident,
				ClosePos: closePos,
			}
		}

		// TODO: this might allow tags inside?
		body := p.parseTagStmtList()
		closePos := p.expect2(token.GTR)

		return &ast.OpenTagStmt{
			OpenPos:  openPos,
			Name:     ident,
			Body:     body,
			ClosePos: closePos,
		}
	case token.ILLEGAL:
		if p.lit == "@" {
			startPos := p.pos
			if len(p.errors) == 0 || p.errors[len(p.errors)-1].Pos != p.file.Position(p.pos) {
				panic(`p.errors does not contain error from the tokenizer caused by the INVALID token ("@")`)
			}
			p.errors = p.errors[:len(p.errors)-1]

			p.next()
			ident := p.parseIdent()

			if p.tok == token.ASSIGN {
				assignPos := p.pos

				p.scanner.AllowTemplateLiteral()
				p.next()

				var val ast.Expr
				if p.tok == token.STRING {
					val = &ast.BasicLit{
						ValuePos: p.pos,
						Kind:     p.tok,
						Value:    p.lit,
					}
					p.next()
				} else if p.tok == token.STRING_TEMPLATE {
					val = p.parseTemplateLiteral()
					p.next()
				} else {
					p.expect(token.STRING)
				}

				return &ast.AttributeStmt{
					StartPos:  startPos,
					AttrName:  ident,
					AssignPos: assignPos,
					Value:     val,
					EndPos:    p.pos,
				}
			}

			return &ast.AttributeStmt{
				StartPos: startPos,
				AttrName: ident,
				EndPos:   ident.Pos() - 1,
			}
		}
	}

	return nil
}

func (p *parser) parseTagStmtList() (list []ast.Stmt) {
	if p.trace {
		defer un(trace(p, "TagStatementList"))
	}

	for p.tok != token.CASE && p.tok != token.DEFAULT && p.tok != token.GTR && p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseStmt())
	}

	return
}

func (p *parser) parseTemplateLiteral() *ast.TemplateLiteralExpr {
	startPos := p.pos
	strings := []string{p.lit}
	parts := []ast.Expr{}

	var closePos token.Pos

	for {
		p.next()
		parts = append(parts, p.parseExpr())
		if p.tok != token.RBRACE {
			p.errorExpected(p.pos, "'"+token.RBRACE.String()+"'")
		}
		p.pos, p.tok, p.lit = p.scanner.TemplateLiteralContinue()
		strings = append(strings, p.lit)
		if p.tok == token.STRING {
			closePos = p.pos
			break
		}
	}

	p.next()

	return &ast.TemplateLiteralExpr{
		OpenPos:  startPos,
		Strings:  strings,
		Parts:    parts,
		ClosePos: closePos,
	}
}
