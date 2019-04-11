// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package code

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const (
	packagePath     = "github.com/pingcap/failpoint"
	packageName     = "failpoint"
	evalFunction    = "Eval"
	evalCtxFunction = "EvalContext"
	extendPkgName   = "_curpkg_"
)

type Rewriter struct {
	rewriteDir    string
	currentPath   string
	currentFile   *ast.File
	failpointName string
	rewritten     bool
}

func NewRewriter(path string) *Rewriter {
	return &Rewriter{
		rewriteDir: path,
	}
}

func (r *Rewriter) rewriteFuncLit(fn *ast.FuncLit) error {
	return r.rewriteStmts(fn.Body.List)
}

func (r *Rewriter) rewriteCallExpr(call *ast.CallExpr) error {
	if fn, ok := call.Fun.(*ast.FuncLit); ok {
		err := r.rewriteFuncLit(fn)
		if err != nil {
			return err
		}
	}

	for _, arg := range call.Args {
		if fn, ok := arg.(*ast.FuncLit); ok {
			err := r.rewriteFuncLit(fn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Rewriter) rewriteAssign(v *ast.AssignStmt) error {
	// fn1, fn2, fn3, ... := func(){...}, func(){...}, func(){...}, ...
	// x, fn := 100, func() {
	//     failpoint.Marker(fpname, func() {
	//         ...
	//     })
	// }
	// ch := <-func() chan interface{} {
	//     failpoint.Marker(fpname, func() {
	//         ...
	//     })
	// }
	for _, v := range v.Rhs {
		if fn, ok := v.(*ast.FuncLit); ok {
			err := r.rewriteFuncLit(fn)
			if err != nil {
				return err
			}
		}
		if call, ok := v.(*ast.CallExpr); ok {
			err := r.rewriteCallExpr(call)
			if err != nil {
				return err
			}
		}
		if sendOrRecv, ok := v.(*ast.UnaryExpr); ok {
			if callExpr, ok2 := sendOrRecv.X.(*ast.CallExpr); ok2 {
				err := r.rewriteCallExpr(callExpr)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Rewriter) rewriteBinaryExpr(expr *ast.BinaryExpr) error {
	// a && func() bool {...} ()
	// func() bool {...} () && a
	if fn, ok := expr.X.(*ast.CallExpr); ok {
		err := r.rewriteCallExpr(fn)
		if err != nil {
			return err
		}
	}
	if fn, ok := expr.Y.(*ast.CallExpr); ok {
		err := r.rewriteCallExpr(fn)
		if err != nil {
			return err
		}
	}
	// func() bool {...} () && func() bool {...} () && a
	// func() bool {...} () && a && func() bool {...} () && a
	if binaryExpr, ok := expr.X.(*ast.BinaryExpr); ok {
		err := r.rewriteBinaryExpr(binaryExpr)
		if err != nil {
			return err
		}
	}
	if binaryExpr, ok := expr.Y.(*ast.BinaryExpr); ok {
		err := r.rewriteBinaryExpr(binaryExpr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Rewriter) rewriteIfStmt(v *ast.IfStmt) error {
	// if a, b := func() {...}, func() int {...}(); cond {...}
	if v.Init != nil {
		err := r.rewriteAssign(v.Init.(*ast.AssignStmt))
		if err != nil {
			return err
		}
	}
	if binaryExpr, ok := v.Cond.(*ast.BinaryExpr); ok {
		err := r.rewriteBinaryExpr(binaryExpr)
		if err != nil {
			return err
		}
	}
	err := r.rewriteStmts(v.Body.List)
	if err != nil {
		return err
	}
	if v.Else != nil {
		if elseIf, ok := v.Else.(*ast.IfStmt); ok {
			return r.rewriteIfStmt(elseIf)
		}
		if els, ok := v.Else.(*ast.BlockStmt); ok {
			return r.rewriteStmts(els.List)
		}
	}
	return nil
}

func (r *Rewriter) rewriteExprs(exprs []ast.Expr) error {
	for _, expr := range exprs {
		// return func(){...},
		if fn, ok := expr.(*ast.FuncLit); ok {
			err := r.rewriteFuncLit(fn)
			if err != nil {
				return err
			}
		}
		// return func() int {...}()
		if fn, ok := expr.(*ast.CallExpr); ok {
			err := r.rewriteCallExpr(fn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Rewriter) rewriteStmts(stmts []ast.Stmt) error {
	for i, block := range stmts {
		switch v := block.(type) {
		case *ast.DeclStmt:
			// var fn1, fn2, fn3, ... = func(){...}, func(){...}, func(){...}, ...
			// var x, fn = 100, func() {
			//     failpoint.Marker(fpname, func() {
			//         ...
			//     })
			// }
			specs := v.Decl.(*ast.GenDecl).Specs
			for _, spec := range specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, v := range vs.Values {
					fn, ok := v.(*ast.FuncLit)
					if !ok {
						continue
					}
					err := r.rewriteStmts(fn.Body.List)
					if err != nil {
						return err
					}
				}
			}

		case *ast.ExprStmt:
			// failpoint.Marker("failpoint.name", func(context.Context, *failpoint.Arg)) {...}
			// failpoint.Break()
			// failpoint.Break("label")
			// failpoint.Continue()
			// failpoint.Fallthrough()
			// failpoint.Continue("label")
			// failpoint.Goto("label")
			// failpoint.Label("label")
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
			if !rewritten {
				continue
			}

			if ifStmt, ok := stmt.(*ast.IfStmt); ok {
				err := r.rewriteIfStmt(ifStmt)
				if err != nil {
					return err
				}
			}

			stmts[i] = stmt
			r.rewritten = true

		case *ast.AssignStmt:
			err := r.rewriteAssign(v)
			if err != nil {
				return err
			}

		case *ast.GoStmt:
			// go func() {...}()
			// go func(fn) {...}(func(){...})
			err := r.rewriteCallExpr(v.Call)
			if err != nil {
				return err
			}

		case *ast.DeferStmt:
			// defer func() {...}()
			// defer func(fn) {...}(func(){...})
			err := r.rewriteCallExpr(v.Call)
			if err != nil {
				return err
			}

		case *ast.ReturnStmt:
			err := r.rewriteExprs(v.Results)
			if err != nil {
				return err
			}

		case *ast.BlockStmt:
			err := r.rewriteStmts(v.List)
			if err != nil {
				return err
			}

		case *ast.IfStmt:
			err := r.rewriteIfStmt(v)
			if err != nil {
				return err
			}

		case *ast.CaseClause:
			// case func() int {...}() > 100 && func () bool {...}()
			if len(v.List) > 0 {
				err := r.rewriteExprs(v.List)
				if err != nil {
					return err
				}
			}
			// case func() int {...}() > 100 && func () bool {...}():
			//     fn := func(){...}
			//     fn()
			if len(v.Body) > 0 {
				err := r.rewriteStmts(v.Body)
				if err != nil {
					return err
				}
			}

		case *ast.SwitchStmt:
			if v.Init != nil {
				err := r.rewriteAssign(v.Init.(*ast.AssignStmt))
				if err != nil {
					return err
				}
			}
			if binaryExpr, ok := v.Tag.(*ast.BinaryExpr); ok {
				err := r.rewriteBinaryExpr(binaryExpr)
				if err != nil {
					return err
				}
			} else if callExpr, ok := v.Tag.(*ast.CallExpr); ok {
				err := r.rewriteCallExpr(callExpr)
				if err != nil {
					return err
				}
			}
			err := r.rewriteStmts(v.Body.List)
			if err != nil {
				return err
			}

		case *ast.CommClause:
			// select {
			// case ch := <-func() chan bool {...}():
			// case <- fromCh:
			// case toCh <- x:
			// case <- func() chan bool {...}():
			// default:
			// }
			if v.Comm != nil {
				if assign, ok := v.Comm.(*ast.AssignStmt); ok {
					err := r.rewriteAssign(assign)
					if err != nil {
						return err
					}
				}
				if expr, ok := v.Comm.(*ast.ExprStmt); ok {
					sendOrRecv := expr.X.(*ast.UnaryExpr)
					if callExpr, ok2 := sendOrRecv.X.(*ast.CallExpr); ok2 {
						err := r.rewriteCallExpr(callExpr)
						if err != nil {
							return err
						}
					}
				}
			}
			err := r.rewriteStmts(v.Body)
			if err != nil {
				return err
			}

		case *ast.SelectStmt:
			if len(v.Body.List) < 1 {
				continue
			}
			err := r.rewriteStmts(v.Body.List)
			if err != nil {
				return err
			}

		case *ast.ForStmt:
			// for i := func() int {...}(); i < func() int {...}(); i += func() int {...}() {...}
			if v.Init != nil {
				err := r.rewriteAssign(v.Init.(*ast.AssignStmt))
				if err != nil {
					return err
				}
			}
			if v.Cond != nil {
				err := r.rewriteBinaryExpr(v.Cond.(*ast.BinaryExpr))
				if err != nil {
					return err
				}
			}
			if v.Post != nil {
				assign, ok := v.Post.(*ast.AssignStmt)
				if ok {
					err := r.rewriteAssign(assign)
					if err != nil {
						return err
					}
				}
			}
			err := r.rewriteStmts(v.Body.List)
			if err != nil {
				return err
			}

		case *ast.RangeStmt:
			if callExpr, ok := v.X.(*ast.CallExpr); ok {
				err := r.rewriteCallExpr(callExpr)
				if err != nil {
					return err
				}
			}
			err := r.rewriteStmts(v.Body.List)
			if err != nil {
				return err
			}

		default:
			fmt.Printf("unsupport statement: %T\n", v)
		}
	}

	// Label statement must ahead of for loop
	for i := 0; i < len(stmts); i++ {
		stmt := stmts[i]
		if label, ok := stmt.(*ast.LabeledStmt); ok {
			label.Stmt = stmts[i+1]
			stmts[i+1] = &ast.EmptyStmt{}
		}
	}
	return nil
}

func (r *Rewriter) rewriteFuncDecl(fn *ast.FuncDecl) error {
	return r.rewriteStmts(fn.Body.List)
}

func (r *Rewriter) rewriteFile(path string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v\n%s", e, debug.Stack())
		}
	}()
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

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if err := r.rewriteFuncDecl(fn); err != nil {
			return err
		}
	}

	if !r.rewritten {
		return nil
	}

	// Generate binding code
	found, err := isBindingFileExists(path)
	if err != nil {
		return err
	}
	if !found {
		err := writeBindingFile(path, file.Name.Name)
		if err != nil {
			return err
		}
	}

	// Backup origin file and replace content
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
		if strings.HasSuffix(path, failpointBindingFileName) {
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
