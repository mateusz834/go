package parser

import (
	"go/ast"
	"go/scanner"
	"go/token"
	"testing"
)

const tgosrc = `package main

import "github.com/mateusz834/tgo"

func test(ctx *tgo.Context, sth string) error {
	<div>"RTFM at google with \{sth+sth} \{sth}"</div>
	//c := a {
	//	"aa \{xd}",
	//}
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
