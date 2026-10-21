package main

import (
 "bytes"
 "encoding/json"
 "go/ast"
 "go/format"
 "go/parser"
 "go/token"
 "os"
 "path/filepath"
 "strconv"
 "strings"
)

type Row struct {
 Kind string `json:"kind"`
 Symbol string `json:"symbol,omitempty"`
 Parent string `json:"parent,omitempty"`
 Name string `json:"name,omitempty"`
 Use string `json:"use,omitempty"`
 File string `json:"file"`
 Line int `json:"line"`
 Runnable bool `json:"runnable,omitempty"`
 Value string `json:"value,omitempty"`
}

func main() {
 root := os.Args[1]
 fs := token.NewFileSet()
 rows := []Row{}
 files, _ := filepath.Glob(filepath.Join(root, "cmd/bd/*.go"))
 expr := func(n ast.Node) string {var b bytes.Buffer; format.Node(&b,fs,n); return b.String()}
 str := func(n ast.Expr) string {
  if l,ok:=n.(*ast.BasicLit); ok && l.Kind==token.STRING {s,_:=strconv.Unquote(l.Value); return s}
  return ""
 }
 for _, f := range files {
  n,e:=parser.ParseFile(fs,f,nil,0); if e!=nil {panic(e)}
  rel,_:=filepath.Rel(root,f)
  istest:=strings.HasSuffix(f,"_test.go")
  add:=func(r Row,p token.Pos){r.File=rel; r.Line=fs.Position(p).Line; rows=append(rows,r)}
  ast.Inspect(n,func(node ast.Node) bool {
   switch x:=node.(type) {
   case *ast.FuncDecl:
    if istest && strings.HasPrefix(x.Name.Name,"Test") {add(Row{Kind:"test",Name:x.Name.Name},x.Pos())}
   case *ast.ValueSpec:
    if istest {return true}
    for i,v:=range x.Values {
     if i>=len(x.Names) {continue}
     name:=x.Names[i].Name
     if u,ok:=v.(*ast.UnaryExpr);ok {v=u.X}
     c,ok:=v.(*ast.CompositeLit);if !ok {continue}
     if expr(c.Type)=="cobra.Command" {
      r:=Row{Kind:"command",Symbol:name}
      for _,elt:=range c.Elts {
       kv,ok:=elt.(*ast.KeyValueExpr);if !ok {continue}
       switch expr(kv.Key) {
       case "Use":r.Use=str(kv.Value);if words:=strings.Fields(r.Use);len(words)>0 {r.Name=words[0]}
       case "Run","RunE":r.Runnable=true
       }
      }
      add(r,c.Pos())
     }
     if name=="proxyPermittedPaths" {
      for _,elt:=range c.Elts {if v,ok:=elt.(ast.Expr);ok {add(Row{Kind:"registry",Name:str(v)},v.Pos())}}
     }
    }
   case *ast.CallExpr:
    if istest {return true}
    sel,ok:=x.Fun.(*ast.SelectorExpr)
    if ok && sel.Sel.Name=="AddCommand" {
     for _,a:=range x.Args {add(Row{Kind:"edge",Parent:expr(sel.X),Symbol:expr(a)},x.Pos())}
    }
    if ok && len(x.Args)>0 {
     receiver:=expr(sel.X)
     if (strings.HasSuffix(receiver,".Flags()")||strings.HasSuffix(receiver,".PersistentFlags()")) && !strings.HasPrefix(sel.Sel.Name,"Mark") {
      name:=str(x.Args[0]);if name==""&&len(x.Args)>1 {name=str(x.Args[1])}
      add(Row{Kind:"flag",Parent:receiver,Name:name,Value:sel.Sel.Name},x.Pos())
     }
    }
    id,ok:=x.Fun.(*ast.Ident)
    if ok && len(x.Args)>0 {
     switch id.Name {
     case "refusedPath","refusedArgs","backupRefusal","permitted","refusedInRunE":
      name:=str(x.Args[0]);if name!="" {
       r:=Row{Kind:"registry",Name:name,Value:id.Name}
       if id.Name=="refusedArgs" {r.Value+=" "+str(x.Args[1])}
       add(r,x.Pos())
      }
     }
    }
   }
   return true
  })
 }
 json.NewEncoder(os.Stdout).Encode(rows)
}
