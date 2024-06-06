package parser

import (
	"fmt"
	"go/ast"
	"go/token"
)

func (p *parser) nextTgoTemplate() {
	if p.tok == token.STRING_TEMPLATE {
		pos := p.pos
		p.templateLit = append(p.templateLit, nil)
		i := len(p.templateLit) - 1
		p.templateLit[i] = p.parseTemplateLiteral()
		p.tok = token.STRING_TEMPLATE
		p.pos = pos
		p.lit = ""
	}
}

func (p *parser) parseTgoStmt() (s ast.Stmt) {
	switch p.tok {
	case token.STRING_TEMPLATE:
		p.next()
		p.expectSemi()
		lit := p.templateLit[len(p.templateLit)-1]
		p.templateLit = p.templateLit[:len(p.templateLit)-1]
		if lit == nil {
			// TODO: figure out if this can happen
			panic("unreachable")
		}
		return &ast.ExprStmt{X: lit}
	case token.LSS, token.END_TAG:
		openPos := p.pos
		closing := p.tok == token.END_TAG
		p.next()

		ident := p.parseIdent()

		if p.tok != token.AT && p.tok != token.GTR {
			p.expectSemi()
		}

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
	case token.AT:
		startPos := p.pos

		p.next()
		ident := p.parseIdent()

		if p.tok == token.ASSIGN {
			assignPos := p.pos

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
				lit := p.templateLit[len(p.templateLit)-1]
				p.templateLit = p.templateLit[:len(p.templateLit)-1]
				if lit == nil {
					panic("unreachable")
				}
				val = lit
				p.next()
			} else {
				p.expect(token.STRING)
			}

			endPos := assignPos
			if val != nil {
				endPos = val.End() - 1
			}

			if p.tok != token.AT && p.tok != token.GTR {
				p.expectSemi()
			}

			return &ast.AttributeStmt{
				StartPos:  startPos,
				AttrName:  ident,
				AssignPos: assignPos,
				Value:     val,
				EndPos:    endPos,
			}
		}

		if p.tok != token.AT && p.tok != token.GTR {
			p.expectSemi()
		}

		return &ast.AttributeStmt{
			StartPos: startPos,
			AttrName: ident,
			EndPos:   ident.End() - 1,
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

	return &ast.TemplateLiteralExpr{
		OpenPos:  startPos,
		Strings:  strings,
		Parts:    parts,
		ClosePos: closePos,
	}
}

type AnalyzeError struct {
	Message          string
	StartPos, EndPos token.Position
}

func (a AnalyzeError) Error() string {
	return fmt.Sprintf("%v: %v", a.StartPos, a.Message)
}

type AnalyzeErrors []AnalyzeError

func (a AnalyzeErrors) Error() string {
	return a[0].Error()
}

type analyzerContext struct {
	errors AnalyzeErrors
	fs     *token.FileSet
}

type context uint8

const (
	contextNotTgo context = iota
	contextTgoBody
	contextTgoTag
)

type analyzer struct {
	context context
	ctx     *analyzerContext
}

func (f *analyzer) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if len(n.Type.Params.List) == 0 {
			return &analyzer{context: contextNotTgo, ctx: f.ctx}
		}
		return &analyzer{context: contextTgoBody, ctx: f.ctx}
	case *ast.FuncLit:
		if len(n.Type.Params.List) == 0 {
			return &analyzer{context: contextNotTgo, ctx: f.ctx}
		}
		return &analyzer{context: contextTgoBody, ctx: f.ctx}
	case *ast.BlockStmt, *ast.IfStmt,
		*ast.SwitchStmt, *ast.CaseClause,
		*ast.ForStmt, *ast.SelectStmt,
		*ast.CommClause, *ast.RangeStmt,
		*ast.TypeSwitchStmt, *ast.ExprStmt:
		return f
	case *ast.TemplateLiteralExpr:
		if f.context != contextTgoBody {
			f.ctx.errors = append(f.ctx.errors, AnalyzeError{
				Message:  "Template literal is not allowed in this context",
				StartPos: f.ctx.fs.Position(n.Pos()),
				EndPos:   f.ctx.fs.Position(n.End()),
			})
		}
		return f
	case *ast.OpenTagStmt:
		if f.context != contextTgoBody {
			f.ctx.errors = append(f.ctx.errors, AnalyzeError{
				Message:  "Open Tag is not allowed in this context",
				StartPos: f.ctx.fs.Position(n.Pos()),
				EndPos:   f.ctx.fs.Position(n.End()),
			})
		}
		return &analyzer{context: contextTgoTag, ctx: f.ctx}
	case *ast.EndTagStmt:
		if f.context != contextTgoBody {
			f.ctx.errors = append(f.ctx.errors, AnalyzeError{
				Message:  "End Tag is not allowed in this context",
				StartPos: f.ctx.fs.Position(n.Pos()),
				EndPos:   f.ctx.fs.Position(n.End()),
			})
		}
		return nil
	case *ast.AttributeStmt:
		if f.context != contextTgoTag {
			f.ctx.errors = append(f.ctx.errors, AnalyzeError{
				Message:  "Attribute is not allowed in this context",
				StartPos: f.ctx.fs.Position(n.Pos()),
				EndPos:   f.ctx.fs.Position(n.End()),
			})
		}
		return nil
	default:
		return &analyzer{context: contextNotTgo, ctx: f.ctx}
	}
}
