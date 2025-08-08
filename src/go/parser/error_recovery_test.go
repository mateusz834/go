package parser

import (
	"flag"
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "")

func TestErrorRecovery(t *testing.T) {
	const testdata = "./testdata/error_recovery"
	files, err := os.ReadDir(testdata)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range files {
		t.Run(tt.Name(), func(t *testing.T) {
			// TODO: support test file being a txtar.

			testFilePath := filepath.Join(testdata, tt.Name())
			src, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Fatal(err)
			}

			const separator = "======\n"
			source, wantAst, _ := strings.Cut(string(src), separator)

			var filter string
			if strings.HasPrefix(wantAst, "filter: ") {
				filter, wantAst, _ = strings.Cut(wantAst, "\n")
				filter = filter[len("filter: "):]
			}

			if !*update && wantAst == "" {
				t.Fatal("missing printed AST part, run with -update to auto-generate")
			}

			// Make sure that that we are able to tokenize the entire source code without any error.
			err = tryTokenize(source)
			if err != nil {
				t.Fatalf("failed to tokenize source: %v", err)
			}

			if *update {
				source, err := insertErrors(source)
				if err != nil {
					t.Fatal(err)
				}

				fset := token.NewFileSet()
				f, _ := ParseFile(fset, "", source, ParseComments|SkipObjectResolution|AllErrors)

				n := any(f)
				if filter != "" {
					var err error
					n, err = runFilter(filter, f)
					if err != nil {
						t.Fatal(err)
					}
				}

				checkErrors2(t, source)

				var astOut strings.Builder
				ast.Fprint(&astOut, fset, n, nil)

				printedAst := regexp.MustCompile(`(?m)^\s*\d+  `).ReplaceAllString(astOut.String(), "") // remove line numbers
				printedAst = regexp.MustCompile(`.  `).ReplaceAllString(printedAst, "   ")              // change ".  " indent to "   "

				out := source + separator
				if filter != "" {
					// TODO: format filter.
					out += "filter: " + filter + "\n"
				}
				out += printedAst

				if err := os.WriteFile(testFilePath, []byte(out), 0660); err != nil {
					t.Fatal(err)
				}
			}

			//fset := token.NewFileSet()
			//f, err := ParseFile(fset, tt.Name(), source, SkipObjectResolution|ParseComments)

			_ = wantAst
		})
	}
}

func checkErrors2(t *testing.T, src string) {
	t.Helper()

	fset := token.NewFileSet()
	_, err := ParseFile(fset, "", src, ParseComments|SkipObjectResolution|AllErrors)
	gotErrs, _ := err.(scanner.ErrorList)

	file := token.NewFileSet().AddFile("", -1, len(src))
	var s scanner.Scanner
	s.Init(file, []byte(src), func(pos token.Position, msg string) {
		panic("unreachable: " + msg + src)
	}, scanner.ScanComments)

	var wantErrs scanner.ErrorList
	var prevAfterErrs []int

	var prev int // position of last non-comment, non-semicolon token
	var here int // position immediately after the token at position prev
outer:
	for {
		pos, tok, lit := s.Scan()
		off := file.Offset(pos)
		switch tok {
		case token.EOF:
			break outer
		case token.COMMENT:
			s := errorRx.FindStringSubmatch(lit)
			if len(s) == 3 {
				if s[1] == "HERE" {
					off = here // position right after the previous token prior to comment
				} else if s[1] == "AFTER" {
					off += len(lit) // end of comment
				} else {
					off = prev // token prior to comment
				}

				prevAfterErrs = append(prevAfterErrs, len(wantErrs))
				wantErrs = append(wantErrs, &scanner.Error{
					Pos: file.Position(file.Pos(off)),
					Msg: s[2],
				})

				if s[1] == "AFTER" {
					for _, v := range prevAfterErrs {
						// Make sure that if there are multiple "AFTER" errors
						// in a row, that such errors point to the latest comment.
						wantErrs[v].Pos = wantErrs[len(wantErrs)-1].Pos
					}
				} else {
					prevAfterErrs = prevAfterErrs[:0]
				}
			}
		default:
			prevAfterErrs = prevAfterErrs[:0]
			prev = off
			tokLength := len(lit)
			if !tok.IsLiteral() {
				tokLength = len(tok.String())
			}
			here = prev + tokLength
		}
	}

	if !slices.EqualFunc(gotErrs, wantErrs, func(x, y *scanner.Error) bool { return *x == *y }) {
		t.Error("difference in errors")
		for _, v := range gotErrs {
			t.Logf("got: %v", v)
		}
		for _, v := range wantErrs {
			t.Logf("want: %v", v)
		}
	}
}

// TODO; fuzz test insertErrors.

func fuzzAddDir(f *testing.F, testdata string) {
	files, err := os.ReadDir(testdata)
	if err != nil {
		f.Fatal(err)
	}
	for _, v := range files {
		if v.IsDir() {
			continue
		}

		testFile := filepath.Join(testdata, v.Name())
		content, err := os.ReadFile(testFile)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(content))
	}
}

func FuzzInsertErrors(f *testing.F) {
	fuzzAddDir(f, ".")
	fuzzAddDir(f, "./testdata")
	f.Fuzz(func(t *testing.T, src string) {
		if testing.Verbose() {
			t.Logf("source:\n%q", src)
		}

		if tryTokenize(src) != nil {
			return
		}

		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic("unreachable: " + msg)
		}, scanner.ScanComments)

		for {
			_, tok, _ := s.Scan()
			if tok == token.COMMENT {
				// TODO: investigate
				return // skip for now
			}
			if tok == token.EOF {
				break
			}
		}

		out, err := insertErrors(src)
		if err != nil {
			t.Fatal(err)
		}

		checkErrors2(t, out)
	})
}

func TestTesting(t *testing.T) {
	src, err := insertErrors("//")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("got: %q", src)
	checkErrors2(t, src)
}

// insertErrors returns a modified src, such that it contatins every error
// reported during parsing of src as comments.
//
// Inserts following ERROR comments:
// - /*ERROR msg*/ - position of the previous token
// - /*ERROR AFTER msg*/ - position right after the previous token
// - /*ERROR HERE msg*/ - end of comment
//
// The input source might already contain ERROR comments, if that error is not
// reported anymore it will be removed.
func insertErrors(src string) (string, error) {
	src = removeErrorComments(src) // TODO: inline

	// Insert a space before every token.
	// TODO: explain why.
	{
		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic("unreachable: " + msg)
		}, scanner.ScanComments)

		lastOff := 0

		var out strings.Builder
		for {
			pos, tok, lit := s.Scan()
			off := file.Offset(pos)

			// Don't include fake spaces before EOF tokens and SEMICOLON tokens causes by EOF.
			if tok == token.EOF || (tok == token.SEMICOLON && lit == "\n" && off == len(src)) {
				break // We can break here since, both cases signal that we have reached the end of file.
			}

			out.WriteString(src[lastOff:off])
			out.WriteString(" ") // fake space
			lastOff = off
		}
		out.WriteString(src[lastOff:])
		src = out.String()
	}

	// Insert ERROR comments
	{
		fset := token.NewFileSet()
		f, err := ParseFile(fset, "", src, SkipObjectResolution|ParseComments|AllErrors)
		errs, _ := err.(scanner.ErrorList)

		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic("unreachable: " + msg)
		}, scanner.ScanComments)

		var out strings.Builder
		lastOff := 0

		var prev int // position of last non-comment, non-semicolon token
		var here int // position immediately after the token at position prev

		prevTok := token.ILLEGAL
		for {
			pos, tok, lit := s.Scan()
			off := file.Offset(pos)
			for len(errs) != 0 {
				errOff := errs[0].Pos.Offset
				errMsg := errs[0].Msg
				if (tok == token.EOF || (tok == token.SEMICOLON && lit == "\n" && off == len(src))) && off == errOff {
					out.WriteString(src[lastOff:errOff])
					out.WriteString("/*ERROR AFTER ")
					out.WriteString(errMsg)
					out.WriteString("*/")
					lastOff = errOff
					errs = errs[1:]
				} else if errOff == prev {
					out.WriteString(src[lastOff:here])
					if prevTok == token.QUO {
						out.WriteString(" ")
					}
					out.WriteString("/*ERROR ")
					out.WriteString(errMsg)
					out.WriteString("*/")
					lastOff = here
					errs = errs[1:]
				} else if errOff == here {
					out.WriteString(src[lastOff:errOff])
					if prevTok == token.QUO {
						out.WriteString(" ")
					}
					out.WriteString("/*ERROR HERE ")
					out.WriteString(errMsg)
					out.WriteString("*/")
					out.WriteString(src[errOff:off])
					lastOff = off
					errs = errs[1:]
				} else if errOff > here && errOff < off {
					out.WriteString(src[lastOff:errOff])
					if prevTok == token.QUO {
						out.WriteString(" ")
					}
					out.WriteString("/*ERROR AFTER ")
					out.WriteString(errMsg)
					out.WriteString("*/")
					lastOff = errOff
					errs = errs[1:]
				} else if errOff >= off {
					break // We will place this error next time.
				} else {
					// Reaching here, likely means a bug in the parser, since parser should not report errors
					// with Positions in the middle of a token.
					f := fset.File(f.FileEnd)
					return "", fmt.Errorf("insertErrors: not able to insert error at: %v", f.Position(f.Pos(errOff)))
				}
			}

			prev = off
			tokLength := len(lit)
			if !tok.IsLiteral() {
				tokLength = len(tok.String())
			}
			here = prev + tokLength

			prevTok = tok
			if tok == token.EOF {
				break
			}
		}

		if len(errs) != 0 {
			return "", fmt.Errorf("insertErrors: non-inserted errors left")
		}

		out.WriteString(src[lastOff:])
		src = out.String()
	}

	// Remove artificial spaces, that we have added before.
	{
		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic("unreachable: " + msg)
		}, scanner.ScanComments)

		lastOff := 0

		var out strings.Builder
		for {
			pos, tok, lit := s.Scan()
			off := file.Offset(pos)
			if tok == token.COMMENT && errorRx.MatchString(lit) {
				continue
			}
			if tok == token.EOF || (tok == token.SEMICOLON && lit == "\n" && off == len(src)) {
				break // We can break here since, both cases signal that we have reached the end of file.
			}
			out.WriteString(src[lastOff : off-1])
			lastOff = off
		}
		out.WriteString(src[lastOff:])
		src = out.String()
	}

	return src, nil
}

var errorRx = regexp.MustCompile(`/\*ERROR *(HERE|AFTER)? (.*)\*/`)

// removeErrorComments removes all ERROR comments from src.
func removeErrorComments(src string) string {
	file := token.NewFileSet().AddFile("", -1, len(src))

	var s scanner.Scanner
	s.Init(file, []byte(src), func(pos token.Position, msg string) {
		panic("unreachable")
	}, scanner.ScanComments)

	var out strings.Builder

	lastOff := 0
outer:
	for {
		pos, tok, lit := s.Scan()
		off := file.Offset(pos)
		switch tok {
		case token.EOF:
			break outer
		case token.COMMENT:
			if errorRx.MatchString(lit) {
				out.WriteString(src[lastOff:off])
				lastOff = off + len(lit)
			}
		}
	}

	out.WriteString(src[lastOff:])
	return out.String()
}

// tryTokenize runs the scanner over the entire src and reports
// the first error that occurs (if any).
func tryTokenize(src string) error {
	file := token.NewFileSet().AddFile("", -1, len(src))
	var err error

	var s scanner.Scanner
	s.Init(file, []byte(src), func(pos token.Position, msg string) {
		if err == nil {
			err = &scanner.Error{
				Pos: pos,
				Msg: msg,
			}
		}
	}, scanner.ScanComments)

	for {
		_, tok, _ := s.Scan()
		if tok == token.EOF {
			break
		}
	}

	return err
}

func runFilter(filter string, f *ast.File) (any, error) {
	fset := token.NewFileSet()
	expr, err := ParseExprFrom(fset, "", filter, 0)
	if err != nil {
		return nil, err
	}

	// TODO: use go/printer to print the location where the error occurred.

	var elem func(val reflect.Value) reflect.Value
	elem = func(val reflect.Value) reflect.Value {
		if val.Kind() == reflect.Pointer {
			return val.Elem()
		} else if val.Kind() == reflect.Interface {
			return elem(val.Elem())
		}
		return val
	}

	//printFilter := func(ast.Expr) string {
	//	var out strings.Builder
	//	if err := format.Node(&out, fset, expr); err != nil {
	//		panic(err)
	//	}
	//	return out.String()
	//}
	//_ = printFilter // use for error messages.

	selected := reflect.ValueOf(f)
	var execute func(ast.Expr) error
	execute = func(e ast.Expr) error {
		switch e := e.(type) {
		case *ast.Ident:
			o := elem(selected).FieldByName(e.Name)
			if o.Kind() == reflect.Invalid {
				return fmt.Errorf("field %q does not exist in %v", e.Name, selected.Type().String())
			}
			selected = o
			return nil
		case *ast.SelectorExpr:
			if err := execute(e.X); err != nil {
				return err
			}
			o := elem(selected).FieldByName(e.Sel.Name)
			if o.Kind() == reflect.Invalid {
				return fmt.Errorf("field %q does not exist in %v", e.Sel.Name, selected.Type().String())
			}
			selected = o
			return nil
		case *ast.IndexExpr:
			if v, ok := e.Index.(*ast.BasicLit); !ok || v.Kind != token.INT {
				return fmt.Errorf("invalid filter, contains %T", e)
			}
			if err := execute(e.X); err != nil {
				return err
			}

			i, err := strconv.ParseInt(e.Index.(*ast.BasicLit).Value, 10, 64)
			if err != nil {
				panic(err) // AST is valid, should not happen
			}

			// TODO: bounds check
			selected = selected.Index(int(i))
			return nil
		default:
			return fmt.Errorf("invalid filter, contains %T", e)
		}
	}

	if err := execute(expr); err != nil {
		return nil, err
	}
	return selected.Interface(), nil
}
