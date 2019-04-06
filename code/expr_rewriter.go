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
	"go/token"
)

type exprRewriter func(rewriter *Rewriter, call *ast.CallExpr) (rewritten bool, result ast.Stmt, err error)

var exprRewriters = map[string]exprRewriter{
	"Inject":   (*Rewriter).rewriteInject,
	"Inject2":  (*Rewriter).rewriteInject,
	"Break":    (*Rewriter).rewriteBreak,
	"Continue": (*Rewriter).rewriteContinue,
	"Label":    (*Rewriter).rewriteLabel,
	"Goto":     (*Rewriter).rewroteGoto,
}

func (r *Rewriter) rewriteInject(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if len(call.Args) != 2 {
		return false, nil, fmt.Errorf("failpoint: expect 2 arguments but got %v", len(call.Args))
	}
	fpname, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return false, nil, fmt.Errorf("failpoint: first argument expect string literal but got %T", call.Args[0])
	}
	fpbody, ok := call.Args[1].(*ast.FuncLit)
	if !ok {
		return false, nil, fmt.Errorf("failpoint: second argument expect closure but got %T", call.Args[1])
	}
	var body = fpbody.Body.List
	ifBody := &ast.BlockStmt{
		Lbrace: call.Pos(),
		List:   body,
		Rbrace: call.End(),
	}
	var (
		checkCall *ast.CallExpr
		cond      = ast.NewIdent("ok")
		init      *ast.AssignStmt
	)

	if len(fpbody.Type.Params.List) == 2 {
		// closure signature:
		// func(ctx context.Context, val failpoint.Value) {...}
		ctx := fpbody.Type.Params.List[0]
		arg := fpbody.Type.Params.List[1]
		checkCall = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: r.failpointName},
				Sel: &ast.Ident{Name: evalFunction},
			},
			Args: []ast.Expr{fpname},
		}
		if ctx.Names[0].Name != "_" {
			checkCall.Args = append(checkCall.Args, ctx.Names[0])
		}
		init = &ast.AssignStmt{
			Lhs: []ast.Expr{cond, arg.Names[0]},
			Rhs: []ast.Expr{checkCall},
			Tok: token.DEFINE,
		}
	} else {
		// closure signature:
		// func(val failpoint.Value) {...}
		arg := fpbody.Type.Params.List[0]
		checkCall = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: r.failpointName},
				Sel: &ast.Ident{Name: evalFunction},
			},
			Args: []ast.Expr{fpname},
		}
		init = &ast.AssignStmt{
			Lhs: []ast.Expr{cond, arg.Names[0]},
			Rhs: []ast.Expr{checkCall},
			Tok: token.DEFINE,
		}
	}

	stmt := &ast.IfStmt{
		If:   call.Pos(),
		Init: init,
		Cond: cond,
		Body: ifBody,
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteBreak(call *ast.CallExpr) (bool, ast.Stmt, error) {
	panic("not implement")
}

func (r *Rewriter) rewriteContinue(call *ast.CallExpr) (bool, ast.Stmt, error) {
	panic("not implement")
}

func (r *Rewriter) rewriteLabel(call *ast.CallExpr) (bool, ast.Stmt, error) {
	panic("not implement")
}

func (r *Rewriter) rewroteGoto(call *ast.CallExpr) (bool, ast.Stmt, error) {
	panic("not implement")
}
