package main

func _() {
	var
}/*ERROR "expected 'IDENT', found '}'"*/
/*ERROR "expected type, found newline"*//*ERROR AFTER "expected ';', found 'EOF'"*//*ERROR AFTER "expected '}', found 'EOF'"*/======
*ast.File {
   Doc: nil
   Package: 1:1
   Name: *ast.Ident {
      NamePos: 1:9
      Name: "main"
      Obj: nil
   }
   Decls: []ast.Decl (len = 1) {
      0: *ast.FuncDecl {
         Doc: nil
         Recv: nil
         Name: *ast.Ident {
            NamePos: 3:6
            Name: "_"
            Obj: nil
         }
         Type: *ast.FuncType {
            Func: 3:1
            TypeParams: nil
            Params: *ast.FieldList {
               Opening: 3:7
               List: nil
               Closing: 3:8
            }
            Results: nil
         }
         Body: *ast.BlockStmt {
            Lbrace: 3:10
            List: []ast.Stmt (len = 1) {
               0: *ast.DeclStmt {
                  Decl: *ast.GenDecl {
                     Doc: nil
                     TokPos: 4:2
                     Tok: var
                     Lparen: -
                     Specs: []ast.Spec (len = 1) {
                        0: *ast.ValueSpec {
                           Doc: nil
                           Names: []*ast.Ident (len = 1) {
                              0: *ast.Ident {
                                 NamePos: 5:1
                                 Name: "_"
                                 Obj: nil
                              }
                           }
                           Type: *ast.BadExpr {
                              From: 5:41
                              To: 5:41
                           }
                           Values: nil
                           Comment: *ast.CommentGroup {
                              List: []*ast.Comment (len = 1) {
                                 0: *ast.Comment {
                                    Slash: 5:2
                                    Text: "/*ERROR \"expected 'IDENT', found '}'\"*/"
                                 }
                              }
                           }
                        }
                     }
                     Rparen: -
                  }
               }
            }
            Rbrace: -
         }
      }
   }
   FileStart: 1:1
   FileEnd: 6:127
   Scope: nil
   Imports: nil
   Unresolved: nil
   Comments: []*ast.CommentGroup (len = 2) {
      0: *(obj @ 51)
      1: *ast.CommentGroup {
         List: []*ast.Comment (len = 3) {
            0: *ast.Comment {
               Slash: 6:1
               Text: "/*ERROR \"expected type, found newline\"*/"
            }
            1: *ast.Comment {
               Slash: 6:41
               Text: "/*ERROR AFTER \"expected ';', found 'EOF'\"*/"
            }
            2: *ast.Comment {
               Slash: 6:84
               Text: "/*ERROR AFTER \"expected '}', found 'EOF'\"*/"
            }
         }
      }
   }
   GoVersion: ""
}
