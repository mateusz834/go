package test

func _() {
/*ERROR AFTER expected ';', found 'EOF'*//*ERROR AFTER expected '}', found 'EOF'*/======
*ast.File {
   Doc: nil
   Package: 1:1
   Name: *ast.Ident {
      NamePos: 1:9
      Name: "test"
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
            List: nil
            Rbrace: -
         }
      }
   }
   FileStart: 1:1
   FileEnd: 4:83
   Scope: nil
   Imports: nil
   Unresolved: nil
   Comments: []*ast.CommentGroup (len = 1) {
      0: *ast.CommentGroup {
         List: []*ast.Comment (len = 2) {
            0: *ast.Comment {
               Slash: 4:1
               Text: "/*ERROR AFTER expected ';', found 'EOF'*/"
            }
            1: *ast.Comment {
               Slash: 4:42
               Text: "/*ERROR AFTER expected '}', found 'EOF'*/"
            }
         }
      }
   }
   GoVersion: ""
}
