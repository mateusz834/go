package parser

import (
	"flag"
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"internal/diff"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
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

var errorRx = regexp.MustCompile(`^(?s)/\*ERROR(?: (HERE|AFTER|\+\d+))? (.*)\*/$`)

func checkErrors2(t *testing.T, src string /*reportErrs func(src string) scanner.ErrorList*/) {
	t.Helper()
	//if reportErrs == nil {
	//	reportErrs = func(src string) scanner.ErrorList {
	//		_, err := ParseFile(token.NewFileSet(), "", src, SkipObjectResolution|ParseComments|AllErrors)
	//		errs, _ := err.(scanner.ErrorList)
	//		return errs
	//	}
	//}

	_, err := ParseFile(token.NewFileSet(), "", src, SkipObjectResolution|ParseComments|AllErrors)
	gotErrs, _ := err.(scanner.ErrorList)

	file := token.NewFileSet().AddFile("", -1, len(src))
	var s scanner.Scanner
	s.Init(file, []byte(src), func(pos token.Position, msg string) {
		panic(fmt.Sprintf("unreachable %v %q", msg, src))
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
				switch s[1] {
				case "HERE":
					off = here // position right after the previous token prior to comment
				case "AFTER":
					off += len(lit) // end of comment
				default:
					off = prev // token prior to comment
					if strings.HasPrefix(s[1], "+") {
						num, err := strconv.ParseInt(s[1][1:], 10, 64)
						if err != nil {
							t.Fatalf("malformed error comment: %v", err)
						}
						off += int(num)
					}
				}

				msg, err := strconv.Unquote(s[2])
				if err != nil {
					t.Fatalf("malformed error comment: %v", err)
				}

				prevAfterErrs = append(prevAfterErrs, len(wantErrs))
				wantErrs = append(wantErrs, &scanner.Error{
					Pos: file.Position(file.Pos(off)),
					Msg: msg,
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
			if !tok.IsLiteral() && tok != token.COMMENT {
				tokLength = len(tok.String())
			}
			here = prev + tokLength
		}
	}

	if !slices.EqualFunc(gotErrs, wantErrs, func(x, y *scanner.Error) bool { return *x == *y }) {
		// TODO: print as strings and to a diff.
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

func FuzzInsertErrors(f *testing.F) {
	fuzzAddDir := func(f *testing.F, testdata string) {
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

	fuzzAddDir(f, ".")
	fuzzAddDir(f, "./testdata")
	f.Fuzz(func(t *testing.T, src string) {
		if testing.Verbose() {
			t.Logf("source:\n%q", src)
		}

		//type parseError struct {
		//	tokenNum int
		//	tokenOff int
		//	message  string
		//}

		//// Parse errors
		//// In a format:
		//// "TokenNum:TokenOff message"
		//var parseErrors []parseError
		//{
		//	for line := range strings.SplitSeq(errors, "\n") {
		//		pos, msg, ok := strings.Cut(line, " ")
		//		if !ok {
		//			return // invalid error list
		//		}

		//		tokenNumStr, tokenOffStr, ok := strings.Cut(pos, ":")
		//		if !ok {
		//			return // invalid error list
		//		}

		//		tokenNum, err := strconv.ParseUint(tokenNumStr, 10, 31)
		//		if err != nil {
		//			return // invalid error list
		//		}
		//		tokenOff, err := strconv.ParseUint(tokenOffStr, 10, 31)
		//		if err != nil {
		//			return // invalid error list
		//		}

		//		parseErrors = append(parseErrors, parseError{
		//			tokenNum: int(tokenNum),
		//			tokenOff: int(tokenOff),
		//			message:  msg,
		//		})

		//		slices.SortStableFunc(parseErrors, func(x, y parseError) int {
		//			return cmp.Or(
		//				cmp.Compare(x.tokenNum, y.tokenNum),
		//				cmp.Compare(x.tokenOff, y.tokenOff),
		//			)
		//		})
		//	}
		//}

		//fakeParseFile := func(src string) scanner.ErrorList {
		//	parseErrors := slices.Clone(parseErrors)

		//	// Parser can place an error at:
		//	// - Any position within a token (also including impliedSemi SEMICOLON tokens).
		//	// - Directly after a token.

		//	var out scanner.ErrorList
		//	file := token.NewFileSet().AddFile("", -1, len(src))
		//	var s scanner.Scanner
		//	s.Init(file, []byte(src), func(pos token.Position, msg string) {
		//		panic(fmt.Sprintf("unreachable %q", src))
		//	}, scanner.ScanComments)

		//	for tokNum := 0; true; tokNum++ {
		//		pos, tok, lit := s.Scan()
		//		off := file.Offset(pos)

		//		tokLength := len(lit)
		//		if !tok.IsLiteral() && tok != token.COMMENT && tok != token.EOF {
		//			tokLength = len(tok.String())
		//		}

		//		for len(parseErrors) != 0 {
		//			e := parseErrors[0]
		//			if e.tokenNum == tokNum {
		//				if e.tokenOff <= tokLength {
		//					out = append(out, &scanner.Error{
		//						Pos: file.Position(file.Pos(off + e.tokenOff)),
		//						Msg: e.message,
		//					})
		//					parseErrors = parseErrors[1:]
		//				} else {
		//					t.Skip()
		//				}
		//			}
		//		}

		//		if tok == token.EOF {
		//			break
		//		}
		//	}

		//	if len(parseErrors) != 0 {
		//		t.Skip()
		//	}

		//	return out
		//}

		out, err := insertErrors(src)
		if err != nil {
			return
		}

		if testing.Verbose() {
			t.Logf("inserted: %q", out)
		}

		checkErrors2(t, out)

		out2, err := insertErrors(out)
		if err != nil {
			t.Fatal(err)
		}

		if out2 != out {
			t.Errorf("not idempotent")
			diff := diff.Diff("insertErrors(src)", []byte(out), "insertErrors(insertErrors(src))", []byte(out2))
			t.Errorf("diff:\n%s", diff)
		}
	})
}

func TestTesting(t *testing.T) {
	//src, err := insertErrors("`\n`0") // TODO
	//src, err := insertErrors("package A; func a(a){defer 0")
	//src, err := insertErrors(`package A;func _(){defer A.type}`)
	src, err := insertErrors(" ")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("got: %q", src)
	checkErrors2(t, src)

	src2, err := insertErrors(src)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("got2: %q", src2)

	if src != src2 {
		t.Errorf("not idempotent")
		diff := diff.Diff("insertErrors(src)", []byte(src), "insertErrors(insertErrors(src))", []byte(src2))
		t.Errorf("diff:\n%s", diff)
	}
}

// insertErrors returns a modified src, such that it contatins every error
// reported during parsing of src as comments.
//
// Inserts following ERROR comments:
// - /*ERROR msg*/ - position of the previous token
// - /*ERROR +1 msg*/ - position of the previous token plus the specified offset.
// - /*ERROR AFTER msg*/ - end of comment
// - /*ERROR HERE msg*/ - position right after the previous token
//
// The input source might already contain ERROR comments, if an error is not
// reported anymore it will be removed.
func insertErrors(src string /*, reportErrs func(src string) scanner.ErrorList*/) (string, error) {
	//if reportErrs == nil {
	//	reportErrs = func(src string) scanner.ErrorList {
	//		_, err := ParseFile(token.NewFileSet(), "", src, SkipObjectResolution|ParseComments|AllErrors)
	//		errs, _ := err.(scanner.ErrorList)
	//		return errs
	//	}
	//}

	// Check run, make sure that the source can be tokenized without any error.
	{
		// It is not possible to calculate the position of the end token reliabely.
		// See https://go.dev/issue/74958
		if strings.ContainsRune(src, '\r') {
			return "", fmt.Errorf(`insertErrors: not able to handle source code with "\r"`)
		}

		var scanErr error
		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			if scanErr == nil {
				scanErr = &scanner.Error{
					Pos: pos,
					Msg: msg,
				}
			}
		}, scanner.ScanComments)

		for {
			_, tok, lit := s.Scan()
			if tok == token.EOF {
				break
			}

			// TODO: try to handle them.
			if tok == token.COMMENT && (strings.HasPrefix(lit, "//") || strings.Contains(lit, "\n")) {
				return "", fmt.Errorf("insertErrors: not able to handle comment: %q", lit)
			}

			// TODO: maybe enforce that on the error messages instead?
			// TODO: we could somehow escape such "/*" in errors?
			if tok == token.STRING && strings.Contains(lit, "*/") {
				return "", fmt.Errorf(`insertErrors: not able to handle string literals containing "/*": %q`, lit)
			}
		}

		if scanErr != nil {
			return "", scanErr
		}
	}

	// Remove existing ERROR comments.
	{
		file := token.NewFileSet().AddFile("", -1, len(src))

		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic("unreachable")
		}, scanner.ScanComments)

		var out strings.Builder

		lastOff := 0
		prevTok := token.ILLEGAL
	outer:
		for {
			pos, tok, lit := s.Scan()
			off := file.Offset(pos)
			switch tok {
			case token.EOF:
				break outer
			case token.COMMENT:
				if errorRx.MatchString(lit) {
					if prevTok == token.QUO && src[off-1] == ' ' {
						// Make sure insertErrors is idempotent.
						// Skip space, this is likely ours.
						out.WriteString(src[lastOff : off-1])
					} else {
						out.WriteString(src[lastOff:off])
					}
					lastOff = off + len(lit)
				}
			}
			prevTok = tok
		}

		out.WriteString(src[lastOff:])
		src = out.String()
	}

	// Remove BOM (if exists).
	r, size := utf8.DecodeRuneInString(src)
	if r == '\ufeff' {
		src = src[size:]
	}

	// Insert a space before every token.
	// TODO: explain why.
	{
		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic(fmt.Sprintf("unreachable %q", src))
		}, scanner.ScanComments)

		lastOff := 0

		var out strings.Builder
		for {
			pos, tok, lit := s.Scan()
			off := file.Offset(pos)

			if tok == token.EOF || (tok == token.SEMICOLON && lit == "\n" && off == len(src)) {
				// We can break here since, both cases signal that we have reached the end of file.
				// We do such so we don't end up with two fake spaces in a row.
				break
			}

			out.WriteString(src[lastOff:off])
			out.WriteString(" ") // fake space
			lastOff = off
		}
		out.WriteString(src[lastOff:])
		out.WriteString(" ") // artificial space before EOF
		src = out.String()
	}

	// Parse file and insert ERROR comments.
	{
		_, err := ParseFile(token.NewFileSet(), "", src, SkipObjectResolution|ParseComments|AllErrors)
		errs, _ := err.(scanner.ErrorList)
		if !sort.IsSorted(errs) {
			return "", fmt.Errorf("insertErrors: unsorted list")
		}

		file := token.NewFileSet().AddFile("", -1, len(src))
		var s scanner.Scanner
		s.Init(file, []byte(src), func(pos token.Position, msg string) {
			panic("unreachable: " + msg)
		}, scanner.ScanComments)

		var out strings.Builder
		lastOff := 0

		var prev int // position of last non-comment, non-semicolon token
		var here int // position immediately after the token at position prev

		nextSpace := false
		for {
			pos, tok, lit := s.Scan()
			off := file.Offset(pos)

			// Write entirely the previous token.
			out.WriteString(src[lastOff:min(here, len(src))])
			lastOff = min(here, len(src))

			for len(errs) != 0 {
				errOff := errs[0].Pos.Offset
				errMsg := errs[0].Msg

				if (tok == token.EOF || (tok == token.SEMICOLON && lit == "\n")) && off == errOff {
					out.WriteString(src[lastOff : off-1]) // insert until off, but keep the fake space directly before EOF/EOL.
					lastOff = off - 1

					if nextSpace {
						out.WriteString(" ")
						nextSpace = false
					}
					out.WriteString("/*ERROR AFTER ")
					out.WriteString(strconv.Quote(errMsg))
					out.WriteString("*/")
					errs = errs[1:]
				} else if errOff == prev {
					if nextSpace {
						out.WriteString(" ")
						nextSpace = false
					}
					out.WriteString("/*ERROR ")
					out.WriteString(strconv.Quote(errMsg))
					out.WriteString("*/")
					errs = errs[1:]
				} else if errOff == here {
					if nextSpace {
						out.WriteString(" ")
						nextSpace = false
					}
					out.WriteString("/*ERROR HERE ")
					out.WriteString(strconv.Quote(errMsg))
					out.WriteString("*/")
					errs = errs[1:]
				} else if errOff > prev && errOff < here {
					// It would be nice if the parse did not produce such
					// errors where the position is in the middle of a token.
					// Currently it might produce such errors, for example:
					//	func _(){defer A.type}
					if nextSpace {
						out.WriteString(" ")
						nextSpace = false
					}
					out.WriteString("/*ERROR ")
					out.WriteString("+")
					out.WriteString(strconv.FormatInt(int64(errOff-prev), 10))
					out.WriteString(" ")
					out.WriteString(strconv.Quote(errMsg))
					out.WriteString("*/")
					errs = errs[1:]
				} else if errOff >= off {
					break // We will place this error next time.
				} else {
					// TODO: return as error, since this can be caused by a parser.
					panic("unreachable")
				}
			}

			prev = off
			tokLength := len(lit)
			if !tok.IsLiteral() && tok != token.COMMENT {
				tokLength = len(tok.String())
			}
			here = prev + tokLength

			nextSpace = tok == token.QUO

			if tok == token.EOF {
				break
			}
		}

		if len(errs) != 0 {
			panic("unreachable")
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
		out.WriteString(src[lastOff : len(src)-1]) // remove space before EOF
		src = out.String()
	}

	return src, nil
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
	// TODO: or just use x.End() and print directly.

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
