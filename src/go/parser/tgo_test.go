package parser

import (
	"go/ast"
	"go/scanner"
	"go/token"
	"testing"
)

// TODO: test template literal inside of template literal (yeah!)

const tgosrc = `package main

import "github.com/mateusz834/tgo"

func test(ctx *tgo.Context, sth string) error {
	<div
		@test
		@test="val"
		@test="val \{sth}"
	>
		"hello\{sth}"
		"hello\{sth}"
	</div>

	// TODO: better error handling of cases like:
	_ = "hello \{sth}"

	// TODO: fix
	// <div> "test" </div>
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
