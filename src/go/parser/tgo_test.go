package parser

import (
	"go/ast"
	"go/scanner"
	"go/token"
	"testing"
)

const tgosrc = `package main

func test(sth string) {
	<a
		@class="\{siema()} \{lol}"
		@href="https://google.com/?q=\{sth}"
	>
		"RTFM"
	</a>
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
