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
	"Marker":   (*Rewriter).rewriteMarker,
	"Break":    (*Rewriter).rewriteBreak,
	"Continue": (*Rewriter).rewriteContinue,
	"Label":    (*Rewriter).rewriteLabel,
	"Goto":     (*Rewriter).rewroteGoto,
}

func (r *Rewriter) rewriteMarker(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if len(call.Args) != 2 {
		return false, nil, fmt.Errorf("failpoint.Marker expect 2 arguments but got %v", len(call.Args))
	}
	fpname, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return false, nil, fmt.Errorf("failpoint.Marker first argument expect string literal but got %T", call.Args[0])
	}
	fpbody, ok := call.Args[1].(*ast.FuncLit)
	if !ok {
		return false, nil, fmt.Errorf("failpoint.Marker second argument expect closure but got %T", call.Args[1])
	}
	var body = fpbody.Body.List
	ifBody := &ast.BlockStmt{
		Lbrace: call.Pos(),
		List:   body,
		Rbrace: call.End(),
	}
	// closure signature:
	// func(ctx context.Context, arg *failpoint.Arg) {...}
	ctx := fpbody.Type.Params.List[0]
	arg := fpbody.Type.Params.List[1]
	ctxName := ctx.Names[0]
	// Pass a nil to `failpoint.IsActive` if ignore context.Context in fail point closure
	// func(_ context.Context, arg *failpoint.Arg) {...}
	if ctxName.Name == "_" {
		ctxName = &ast.Ident{Name: "nil"}
	}
	checkCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: r.failpointName},
			Sel: &ast.Ident{Name: isActiveCall},
		},
		Args: []ast.Expr{ctxName, fpname},
	}
	cond := ast.NewIdent("ok")
	init := &ast.AssignStmt{
		Lhs: []ast.Expr{cond, arg.Names[0]},
		Rhs: []ast.Expr{checkCall},
		Tok: token.DEFINE,
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
