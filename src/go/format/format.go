// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package format implements standard formatting of Go source.
//
// Note that formatting of Go source code changes over time, so tools relying on
// consistent formatting should execute a specific version of the gofmt binary
// instead of using this package. That way, the formatting will be stable, and
// the tools won't need to be recompiled each time gofmt changes.
//
// For example, pre-submit checks that use this package directly would behave
// differently depending on what Go version each developer uses, causing the
// check to be inherently fragile.
package format

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"slices"
)

// Keep these in sync with cmd/gofmt/gofmt.go.
const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers

	// printerNormalizeNumbers means to canonicalize number literal prefixes
	// and exponents while printing. See https://golang.org/doc/go1.13#gofmt.
	//
	// This value is defined in go/printer specifically for go/format and cmd/gofmt.
	printerNormalizeNumbers = 1 << 30
)

var config = printer.Config{Mode: printerMode, Tabwidth: tabWidth}

const parserMode = parser.ParseComments | parser.SkipObjectResolution

// Node formats node in canonical gofmt style and writes the result to dst.
//
// The node type must be *[ast.File], *[printer.CommentedNode], [][ast.Decl],
// [][ast.Stmt], or assignment-compatible to [ast.Expr], [ast.Decl], [ast.Spec],
// or [ast.Stmt]. Node does not modify node. Imports are not sorted for
// nodes representing partial source files (for instance, if the node is
// not an *[ast.File] or a *[printer.CommentedNode] not wrapping an *[ast.File]).
//
// The function may return early (before the entire result is written)
// and return a formatting error, for instance due to an incorrect AST.
func Node(dst io.Writer, fset *token.FileSet, node any) error {
	// Determine if we have a complete source file (file != nil).
	var file *ast.File
	var cnode *printer.CommentedNode
	switch n := node.(type) {
	case *ast.File:
		file = n
	case *printer.CommentedNode:
		if f, ok := n.Node.(*ast.File); ok {
			file = f
			cnode = n
		}
	}

	// Sort imports if necessary.
	if file != nil && hasUnsortedImports(file) {
		// Make a copy of the AST because ast.SortImports is destructive.
		var err error
		fset, file, err = cloneFileForSortImports(fset, file)
		if err != nil {
			return err
		}

		ast.SortImports(fset, file)

		// Use new file with sorted imports.
		node = file
		if cnode != nil {
			node = &printer.CommentedNode{Node: file, Comments: cnode.Comments}
		}
	}

	return config.Fprint(dst, fset, node)
}

func cloneFileForSortImports(fset *token.FileSet, f *ast.File) (*token.FileSet, *ast.File, error) {
	fileFromFset := fset.File(f.FileStart)
	if fileFromFset == nil {
		return nil, nil, errors.New("invalid *token.FileSet")
	}

	newFset := token.NewFileSet()
	file := newFset.AddFile(fileFromFset.Name(), fileFromFset.Base(), fileFromFset.Size())
	if !file.SetLines(slices.Clone(fileFromFset.Lines())) {
		return nil, nil, errors.New("invalid *token.FileSet")
	}

	newComments := make([]*ast.CommentGroup, len(f.Comments))
	for i, cg := range f.Comments {
		commentGroup := &ast.CommentGroup{
			List: make([]*ast.Comment, len(cg.List)),
		}
		for j, c := range cg.List {
			clonedComment := *c
			commentGroup.List[j] = &clonedComment
		}
		newComments[i] = commentGroup
	}

	clonedFile := *f
	clonedFile.Comments = newComments
	clonedFile.Decls = slices.Clone(f.Decls)

	if len(clonedFile.Decls) > 0 {
		if importDecl, ok := clonedFile.Decls[0].(*ast.GenDecl); ok && importDecl.Tok == token.IMPORT {
			clonedImportDecl := *importDecl

			if clonedImportDecl.Doc != nil {
				clonedImportDecl.Doc = newComments[slices.Index(f.Comments, importDecl.Doc)]
			}

			clonedImportDecl.Specs = make([]ast.Spec, len(importDecl.Specs))
			for i, v := range importDecl.Specs {
				s, ok := v.(*ast.ImportSpec)
				if !ok {
					return nil, nil, fmt.Errorf("import spec contains unexpected type: %T", v)
				}
				clonedImportSpec := *s
				if clonedImportSpec.Doc != nil {
					clonedImportSpec.Doc = newComments[slices.Index(f.Comments, clonedImportSpec.Doc)]
				}
				if clonedImportSpec.Comment != nil {
					clonedImportSpec.Comment = newComments[slices.Index(f.Comments, clonedImportSpec.Comment)]
				}
				clonedImportDecl.Specs[i] = &clonedImportSpec
			}
			clonedFile.Decls[0] = &clonedImportDecl
		}
	}

	return newFset, &clonedFile, nil
}

// Source formats src in canonical gofmt style and returns the result
// or an (I/O or syntax) error. src is expected to be a syntactically
// correct Go source file, or a list of Go declarations or statements.
//
// If src is a partial source file, the leading and trailing space of src
// is applied to the result (such that it has the same leading and trailing
// space as src), and the result is indented by the same amount as the first
// line of src containing code. Imports are not sorted for partial source files.
func Source(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, sourceAdj, indentAdj, err := parse(fset, "", src, true)
	if err != nil {
		return nil, err
	}

	if sourceAdj == nil {
		// Complete source file.
		// TODO(gri) consider doing this always.
		ast.SortImports(fset, file)
	}

	return format(fset, file, sourceAdj, indentAdj, src, config)
}

func hasUnsortedImports(file *ast.File) bool {
	for _, d := range file.Decls {
		d, ok := d.(*ast.GenDecl)
		if !ok || d.Tok != token.IMPORT {
			// Not an import declaration, so we're done.
			// Imports are always first.
			return false
		}
		if d.Lparen.IsValid() {
			// For now assume all grouped imports are unsorted.
			// TODO(gri) Should check if they are sorted already.
			return true
		}
		// Ungrouped imports are sorted by default.
	}
	return false
}
