package parser

import (
	"go/ast"
	"go/token"
	"testing"
)

const tgosrc = `package main

func test() {
	<div
		@a @a="hello"
		"lol"
	>
}
`

func TestTgo(t *testing.T) {
	fs := token.NewFileSet()
	f, err := ParseFile(fs, "test.tgo", tgosrc, SkipObjectResolution)

	ast.Print(fs, f)

	if err != nil {
		t.Fatal(err)
	}
}
