package code

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const (
	packagePath  = "github.com/pingcap/failpoint"
	packageName  = "failpoint"
	isActiveCall = "IsActive"
)

type Rewriter struct {
	rewriteDir    string
	currentPath   string
	currentFile   *ast.File
	failpointName string
}

func NewRewriter(path string) *Rewriter {
	return &Rewriter{
		rewriteDir: path,
	}
}

func (r *Rewriter) rewriteBlockStmts(stmts []ast.Stmt) error {
	for i, block := range stmts {
		switch v := block.(type) {
		case *ast.AssignStmt:
			// var x = func() {
			//     failpoint.Marker(fpname, func() {
			//         ...
			//     })
			// }
			// TODO

		case *ast.ExprStmt:
			call, ok := v.X.(*ast.CallExpr)
			if !ok {
				break
			}
			selectorExpr, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				break
			}
			packageName, ok := selectorExpr.X.(*ast.Ident)
			if !ok || packageName.Name != r.failpointName {
				break
			}
			exprRewriter, found := exprRewriters[selectorExpr.Sel.Name]
			if !found {
				break
			}
			rewritten, stmt, err := exprRewriter(r, call)
			if err != nil {
				return err
			}
			if rewritten {
				stmts[i] = stmt
			}
		}
	}
	return nil
}

func (r *Rewriter) rewriteFuncDecl(fn *ast.FuncDecl) error {
	fmt.Println("rewrite function", fn.Name.Name)
	return r.rewriteBlockStmts(fn.Body.List)
}

func (r *Rewriter) rewriteFile(path string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	if len(file.Decls) < 1 {
		return nil
	}
	r.currentPath = path
	r.currentFile = file

	var failpointImport *ast.ImportSpec
	for _, imp := range file.Imports {
		if strings.Trim(imp.Path.Value, "`\"") == packagePath {
			failpointImport = imp
			break
		}
	}
	if failpointImport == nil {
		panic("import path should be check before rewrite")
	}
	if failpointImport.Name != nil {
		r.failpointName = failpointImport.Name.Name
	} else {
		r.failpointName = packageName
	}

	fmt.Println(path, r.failpointName)
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if err := r.rewriteFuncDecl(fn); err != nil {
			return err
		}
	}

	targetPath := path + failpointStashFileSuffix
	if err := os.Rename(path, targetPath); err != nil {
		return err
	}

	newFile, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer newFile.Close()
	return printer.Fprint(newFile, fset, file)
}

func (r *Rewriter) Rewrite() error {
	var files []string
	err := filepath.Walk(r.rewriteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Will rewrite a file only if the file has imported "github.com/pingcap/failpoint"
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		if len(file.Imports) < 1 {
			return nil
		}
		for _, imp := range file.Imports {
			// import path maybe in the form of:
			//
			// 1. normal import
			//    - "github.com/pingcap/failpoint"
			//    - `github.com/pingcap/failpoint`
			// 2. ignore import
			//    - _ "github.com/pingcap/failpoint"
			//    - _ `github.com/pingcap/failpoint`
			// 3. alias import
			//    - alias "github.com/pingcap/failpoint"
			//    - alias `github.com/pingcap/failpoint`
			// we should trim '"' or '`' before compare it.
			if strings.Trim(imp.Path.Value, "`\"") == packagePath {
				files = append(files, path)
				break
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, file := range files {
		err := r.rewriteFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}
