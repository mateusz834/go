package main

var _ = []int{
	0/*ERROR AFTER "missing ',' before newline in composite literal"*/
/*ERROR AFTER "expected ';', found 'EOF'"*//*ERROR AFTER "expected '}', found 'EOF'"*/======
*ast.File {
   Doc: nil
   Package: 1:1
   Name: *ast.Ident {
      NamePos: 1:9
      Name: "main"
      Obj: nil
   }
   Decls: []ast.Decl (len = 1) {
      0: *ast.GenDecl {
         Doc: nil
         TokPos: 3:1
         Tok: var
         Lparen: -
         Specs: []ast.Spec (len = 1) {
            0: *ast.ValueSpec {
               Doc: nil
               Names: []*ast.Ident (len = 1) {
                  0: *ast.Ident {
                     NamePos: 3:5
                     Name: "_"
                     Obj: nil
                  }
               }
               Type: nil
               Values: []ast.Expr (len = 1) {
                  0: *ast.CompositeLit {
                     Type: *ast.ArrayType {
                        Lbrack: 3:9
                        Len: nil
                        Elt: *ast.Ident {
                           NamePos: 3:11
                           Name: "int"
                           Obj: nil
                        }
                     }
                     Lbrace: 3:14
                     Elts: []ast.Expr (len = 1) {
                        0: *ast.BasicLit {
                           ValuePos: 4:2
                           Kind: INT
                           Value: "0"
                        }
                     }
                     Rbrace: 5:87
                     Incomplete: false
                  }
               }
               Comment: nil
            }
         }
         Rparen: -
      }
   }
   FileStart: 1:1
   FileEnd: 5:87
   Scope: nil
   Imports: nil
   Unresolved: nil
   Comments: []*ast.CommentGroup (len = 2) {
      0: *ast.CommentGroup {
         List: []*ast.Comment (len = 1) {
            0: *ast.Comment {
               Slash: 4:3
               Text: "/*ERROR AFTER \"missing ',' before newline in composite literal\"*/"
            }
         }
      }
      1: *ast.CommentGroup {
         List: []*ast.Comment (len = 2) {
            0: *ast.Comment {
               Slash: 5:1
               Text: "/*ERROR AFTER \"expected ';', found 'EOF'\"*/"
            }
            1: *ast.Comment {
               Slash: 5:44
               Text: "/*ERROR AFTER \"expected '}', found 'EOF'\"*/"
            }
         }
      }
   }
   GoVersion: ""
}
