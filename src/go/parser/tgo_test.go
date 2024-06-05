package parser

import (
	"errors"
	"go/ast"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TODO: get rid of EmptyStmt in Tag body.
// TODO: figure whether we should require ';' after <div and after attributes @attr
// except before attributes.
// like:
// @test @test (no need for a semi)
// <div @test @test> (no need for a semi)
// @test a := 3 (require semi)
// <div; a := 3 (require semi)

const tgosrc = `package main

import "github.com/mateusz834/tgo"

func test(ctx *tgo.Context, sth string) error {
	<div
		@href="test" @test="hello"
	>
		"test \{sth}"
		"test \{sth}"
	</div>

	//<div>
	//	"test \{sth}"
	//</div>
	//"test \{sth}"

	//"hello\{func() string{
	//	"helloooo \{sth}"
	//}()}"

	// TODO: better error handling of cases like:
	//_ = "hello \{sth}"
}
`

func TestTgo(t *testing.T) {
	fs := token.NewFileSet()
	f, err := ParseFile(fs, "test.tgo", tgosrc, SkipObjectResolution)

	ast.Print(fs, f)

	if err != nil {
		if v, ok := err.(scanner.ErrorList); ok {
			for _, err := range v {
				t.Errorf("%v", err)
			}
		}
		t.Fatalf("%v", err)
	}
}

func TestTgoBasicSyntax(t *testing.T) {
	const prefix = "package main\nfunc test() {"
	off := token.Pos(len(prefix)) + 1

	cases := []struct {
		in  string
		out []ast.Stmt
	}{
		{
			in: `<div>`,
			out: []ast.Stmt{
				&ast.OpenTagStmt{
					OpenPos: off,
					Name: &ast.Ident{
						NamePos: off + 1,
						Name:    "div",
					},
					Body:     nil,
					ClosePos: off + 4,
				},
			},
		},
		{
			in: `</div>`,
			out: []ast.Stmt{
				&ast.EndTagStmt{
					OpenPos: off,
					Name: &ast.Ident{
						NamePos: off + 2,
						Name:    "div",
					},
					ClosePos: off + 5,
				},
			},
		},
		{
			in: `"test"`,
			out: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.BasicLit{
						ValuePos: off,
						Kind:     token.STRING,
						Value:    `"test"`,
					},
				},
			},
		},
		{
			in: `"test \{sth}"`,
			out: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.TemplateLiteralExpr{
						OpenPos: off,
						Strings: []string{
							`"test `,
							`"`,
						},
						Parts: []ast.Expr{
							&ast.Ident{
								NamePos: off + 8,
								Name:    "sth",
							},
						},
						ClosePos: off + 12,
					},
				},
			},
		},
		{
			in: `"test \{sth} \{sth}"`,
			out: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.TemplateLiteralExpr{
						OpenPos: off,
						Strings: []string{
							`"test `,
							` `,
							`"`,
						},
						Parts: []ast.Expr{
							&ast.Ident{
								NamePos: off + 8,
								Name:    "sth",
							},
							&ast.Ident{
								NamePos: off + 15,
								Name:    "sth",
							},
						},
						ClosePos: off + 19,
					},
				},
			},
		},
		{
			in: `@attr`,
			out: []ast.Stmt{
				&ast.AttributeStmt{
					StartPos: off,
					AttrName: &ast.Ident{
						NamePos: off + 1,
						Name:    "attr",
					},
					EndPos: off + 4,
				},
			},
		},
		{
			in: `@attr="test"`,
			out: []ast.Stmt{
				&ast.AttributeStmt{
					StartPos: off,
					AttrName: &ast.Ident{
						NamePos: off + 1,
						Name:    "attr",
					},
					AssignPos: off + 5,
					Value: &ast.BasicLit{
						ValuePos: off + 6,
						Kind:     token.STRING,
						Value:    `"test"`,
					},
					EndPos: off + 11,
				},
			},
		},
		{
			in: `@attr="test \{sth}"`,
			out: []ast.Stmt{
				&ast.AttributeStmt{
					StartPos: off,
					AttrName: &ast.Ident{
						NamePos: off + 1,
						Name:    "attr",
					},
					AssignPos: off + 5,
					Value: &ast.TemplateLiteralExpr{
						OpenPos: off + 6,
						Strings: []string{
							`"test `,
							`"`,
						},
						Parts: []ast.Expr{
							&ast.Ident{
								NamePos: off + 14,
								Name:    "sth",
							},
						},
						ClosePos: off + 18,
					},
					EndPos: off + 18,
				},
			},
		},
		{
			in: `<div></div>`,
			out: []ast.Stmt{
				&ast.OpenTagStmt{
					OpenPos: off,
					Name: &ast.Ident{
						NamePos: off + 1,
						Name:    "div",
					},
					Body:     nil,
					ClosePos: off + 4,
				},
				&ast.EndTagStmt{
					OpenPos: off + 5,
					Name: &ast.Ident{
						NamePos: off + 7,
						Name:    "div",
					},
					ClosePos: off + 10,
				},
			},
		},
		{
			in: `<div>"test"</div>`,
			out: []ast.Stmt{
				&ast.OpenTagStmt{
					OpenPos: off,
					Name: &ast.Ident{
						NamePos: off + 1,
						Name:    "div",
					},
					Body:     nil,
					ClosePos: off + 4,
				},
				&ast.ExprStmt{
					X: &ast.BasicLit{
						ValuePos: off + 5,
						Kind:     token.STRING,
						Value:    `"test"`,
					},
				},
				&ast.EndTagStmt{
					OpenPos: off + 11,
					Name: &ast.Ident{
						NamePos: off + 13,
						Name:    "div",
					},
					ClosePos: off + 16,
				},
			},
		},
		{
			in: `<div>"test \{sth}"</div>`,
			out: []ast.Stmt{
				&ast.OpenTagStmt{
					OpenPos: off,
					Name: &ast.Ident{
						NamePos: off + 1,
						Name:    "div",
					},
					Body:     nil,
					ClosePos: off + 4,
				},
				&ast.ExprStmt{
					X: &ast.TemplateLiteralExpr{
						OpenPos: off + 5,
						Strings: []string{
							`"test `,
							`"`,
						},
						Parts: []ast.Expr{
							&ast.Ident{
								NamePos: off + 13,
								Name:    "sth",
							},
						},
						ClosePos: off + 17,
					},
				},
				&ast.EndTagStmt{
					OpenPos: off + 18,
					Name: &ast.Ident{
						NamePos: off + 20,
						Name:    "div",
					},
					ClosePos: off + 23,
				},
			},
		},
	}

	for _, tt := range cases {
		inStr := prefix + tt.in + "}"

		fs := token.NewFileSet()
		f, err := ParseFile(fs, "test.go", inStr, SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}

		if len(f.Decls) == 0 {
			t.Errorf("missing func decl")
			continue
		}
		fd, ok := f.Decls[0].(*ast.FuncDecl)
		if !ok {
			t.Errorf("f.Decls[0] is not *ast.FuncDecl")
			continue
		}

		expectList := fd.Body.List
		if !reflect.DeepEqual(expectList, tt.out) {
			t.Errorf("unexpected AST for:\n%v", inStr)
		}
	}
}

func TestTgoSyntax(t *testing.T) {
	const testdata = "./testdata/tgo"
	files, err := os.ReadDir(testdata)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range files {
		ext := filepath.Ext(v.Name())
		if ext == ".tgo" {
			testFile := filepath.Join(testdata, v.Name())
			expectFileName := filepath.Join(testdata, v.Name()[:len(v.Name())-len(".tgo")]+".ast")

			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatal(err)
			}

			fs := token.NewFileSet()
			f, err := ParseFile(fs, filepath.Base(testFile), content, SkipObjectResolution)
			if err != nil {
				if v, ok := err.(scanner.ErrorList); ok {
					for _, err := range v {
						t.Errorf("%v", err)
					}
				}
				t.Errorf("Error while parsing file %v: %v", testFile, err)
				continue
			}

			var b strings.Builder
			ast.Fprint(&b, fs, f, ast.NotNilFilter)

			expect, err := os.ReadFile(expectFileName)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if err := os.WriteFile(expectFileName, []byte(b.String()), 06660); err != nil {
						t.Fatal(err)
					}
					continue
				}
				t.Fatal(err)
			}

			got := b.String()
			if string(expect) != got {
				t.Errorf("unexpected in %v", testFile)
			}
		}
	}

}
