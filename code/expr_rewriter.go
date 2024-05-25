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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package code

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type exprRewriter func(rewriter *Rewriter, call *ast.CallExpr) (rewritten bool, result ast.Stmt, err error)

var exprRewriters = map[string]exprRewriter{
	"Inject":        (*Rewriter).rewriteInject,
	"InjectContext": (*Rewriter).rewriteInjectContext,
	"InjectCall":    (*Rewriter).rewriteInjectCall,
	"Break":         (*Rewriter).rewriteBreak,
	"Continue":      (*Rewriter).rewriteContinue,
	"Label":         (*Rewriter).rewriteLabel,
	"Goto":          (*Rewriter).rewriteGoto,
	"Fallthrough":   (*Rewriter).rewriteFallthrough,
	"Return":        (*Rewriter).rewriteReturn,
}

func (r *Rewriter) rewriteInject(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if len(call.Args) != 2 {
		return false, nil, fmt.Errorf("failpoint.Inject: expect 2 arguments but got %v in %s", len(call.Args), r.pos(call.Pos()))
	}
	// First argument need not to be a string literal, any string type stuff is ok.
	// Type safe is convinced by compiler.
	fpname, ok := call.Args[0].(ast.Expr)
	if !ok {
		return false, nil, fmt.Errorf("failpoint.Inject: first argument expect a valid expression in %s", r.pos(call.Pos()))
	}

	// failpoint.Inject("failpoint-name", nil)
	ident, ok := call.Args[1].(*ast.Ident)
	isNilFunc := ok && ident.Name == "nil"

	// failpoint.Inject("failpoint-name", func(){...})
	// failpoint.Inject("failpoint-name", func(val failpoint.Value){...})
	fpbody, isFuncLit := call.Args[1].(*ast.FuncLit)
	if !isNilFunc && !isFuncLit {
		return false, nil, fmt.Errorf("failpoint.Inject: second argument expect closure in %s", r.pos(call.Pos()))
	}
	if isFuncLit {
		if len(fpbody.Type.Params.List) > 1 {
			return false, nil, fmt.Errorf("failpoint.Inject: closure signature illegal in %s", r.pos(call.Pos()))
		}

		if len(fpbody.Type.Params.List) == 1 && len(fpbody.Type.Params.List[0].Names) > 1 {
			return false, nil, fmt.Errorf("failpoint.Inject: closure signature illegal in %s", r.pos(call.Pos()))
		}
	}

	fpnameExtendCall := &ast.CallExpr{
		Fun:  ast.NewIdent(ExtendPkgName),
		Args: []ast.Expr{fpname},
	}

	checkCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{NamePos: call.Pos(), Name: r.failpointName},
			Sel: ast.NewIdent(evalFunction),
		},
		Args: []ast.Expr{fpnameExtendCall},
	}
	if isNilFunc || len(fpbody.Body.List) < 1 {
		return true, &ast.ExprStmt{X: checkCall}, nil
	}

	ifBody := &ast.BlockStmt{
		Lbrace: call.Pos(),
		List:   fpbody.Body.List,
		Rbrace: call.End(),
	}

	// closure signature:
	// func(val failpoint.Value) {...}
	// func() {...}
	var argName *ast.Ident
	if len(fpbody.Type.Params.List) > 0 {
		arg := fpbody.Type.Params.List[0]
		selector, ok := arg.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Value" || selector.X.(*ast.Ident).Name != r.failpointName {
			return false, nil, fmt.Errorf("failpoint.Inject: invalid signature in %s", r.pos(call.Pos()))
		}
		argName = arg.Names[0]
	} else {
		argName = ast.NewIdent("_")
	}

	err := ast.NewIdent("_err_")
	init := &ast.AssignStmt{
		Lhs: []ast.Expr{argName, err},
		Rhs: []ast.Expr{checkCall},
		Tok: token.DEFINE,
	}

	cond := &ast.BinaryExpr{
		X:  err,
		Op: token.EQL,
		Y:  ast.NewIdent("nil"),
	}
	stmt := &ast.IfStmt{
		If:   call.Pos(),
		Init: init,
		Cond: cond,
		Body: ifBody,
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteInjectContext(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if len(call.Args) != 3 {
		return false, nil, fmt.Errorf("failpoint.InjectContext: expect 3 arguments but got %v in %s", len(call.Args), r.pos(call.Pos()))
	}

	// Second argument need not to be a identifier, any context type token (e.g. selector) is OK.
	// Type safe is convinced by compiler.
	ctxname, ok := call.Args[0].(ast.Expr)
	if !ok {
		return false, nil, fmt.Errorf("failpoint.InjectContext: first argument expect context in %s, which must be an expression", r.pos(call.Pos()))
	}
	// Second argument need not to be a string literal, any string type stuff is ok.
	// Type safe is convinced by compiler.
	fpname, ok := call.Args[1].(ast.Expr)
	if !ok {
		return false, nil, fmt.Errorf("failpoint.InjectContext: second argument expect a valid expression in %s", r.pos(call.Pos()))
	}

	// failpoint.InjectContext("failpoint-name", ctx, nil)
	ident, ok := call.Args[2].(*ast.Ident)
	isNilFunc := ok && ident.Name == "nil"

	// failpoint.InjectContext("failpoint-name", ctx, func(){...})
	// failpoint.InjectContext("failpoint-name", ctx, func(val failpoint.Value){...})
	fpbody, isFuncLit := call.Args[2].(*ast.FuncLit)
	if !isNilFunc && !isFuncLit {
		return false, nil, fmt.Errorf("failpoint.InjectContext: third argument expect closure in %s", r.pos(call.Pos()))
	}

	if isFuncLit {
		if len(fpbody.Type.Params.List) > 1 {
			return false, nil, fmt.Errorf("failpoint.InjectContext: closure signature illegal in %s", r.pos(call.Pos()))
		}

		if len(fpbody.Type.Params.List) == 1 && len(fpbody.Type.Params.List[0].Names) > 1 {
			return false, nil, fmt.Errorf("failpoint.InjectContext: closure signature illegal in %s", r.pos(call.Pos()))
		}
	}

	fpnameExtendCall := &ast.CallExpr{
		Fun:  ast.NewIdent(ExtendPkgName),
		Args: []ast.Expr{fpname},
	}

	checkCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{NamePos: call.Pos(), Name: r.failpointName},
			Sel: ast.NewIdent(evalCtxFunction),
		},
		Args: []ast.Expr{ctxname, fpnameExtendCall},
	}
	if isNilFunc || len(fpbody.Body.List) < 1 {
		return true, &ast.ExprStmt{X: checkCall}, nil
	}

	ifBody := &ast.BlockStmt{
		Lbrace: call.Pos(),
		List:   fpbody.Body.List,
		Rbrace: call.End(),
	}

	// closure signature:
	// func(val failpoint.Value) {...}
	// func() {...}
	var argName *ast.Ident
	if len(fpbody.Type.Params.List) > 0 {
		arg := fpbody.Type.Params.List[0]
		selector, ok := arg.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Value" || selector.X.(*ast.Ident).Name != r.failpointName {
			return false, nil, fmt.Errorf("failpoint.InjectContext: invalid signature in %s", r.pos(call.Pos()))
		}
		argName = arg.Names[0]
	} else {
		argName = ast.NewIdent("_")
	}

	err := ast.NewIdent("_err_")
	init := &ast.AssignStmt{
		Lhs: []ast.Expr{argName, err},
		Rhs: []ast.Expr{checkCall},
		Tok: token.DEFINE,
	}

	cond := &ast.BinaryExpr{
		X:  err,
		Op: token.EQL,
		Y:  ast.NewIdent("nil"),
	}
	stmt := &ast.IfStmt{
		If:   call.Pos(),
		Init: init,
		Cond: cond,
		Body: ifBody,
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteInjectCall(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if len(call.Args) < 1 {
		return false, nil, fmt.Errorf("failpoint.InjectCall: expect at least 1 arguments but got %v in %s", len(call.Args), r.pos(call.Pos()))
	}
	// First argument need not to be a string literal, any string type stuff is ok.
	// Type safe is convinced by compiler.
	fpname, ok := call.Args[0].(ast.Expr)
	if !ok {
		return false, nil, fmt.Errorf("failpoint.InjectCall: first argument expect a valid expression in %s", r.pos(call.Pos()))
	}

	fpnameExtendCall := &ast.CallExpr{
		Fun:  ast.NewIdent(ExtendPkgName),
		Args: []ast.Expr{fpname},
	}

	// failpoint.InjectFn("name", a, b, c)
	//    |
	//    v
	// if _, _err_ := failpoint.Eval(_curpkg_("name")); _err_ == nil {
	//     failpoint.Call(_curpkg_("name"), a, b, c)
	// }
	fnArgs := make([]ast.Expr, 0, len(call.Args))
	fnArgs = append(fnArgs, fpnameExtendCall)
	fnArgs = append(fnArgs, call.Args[1:]...)
	fnCall := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{NamePos: call.Pos(), Name: r.failpointName},
				Sel: ast.NewIdent(callFunction),
			},
			Args: fnArgs,
		},
	}
	ifBody := &ast.BlockStmt{
		Lbrace: call.Pos(),
		List:   []ast.Stmt{fnCall},
		Rbrace: call.End(),
	}

	checkCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{NamePos: call.Pos(), Name: r.failpointName},
			Sel: ast.NewIdent(evalFunction),
		},
		Args: []ast.Expr{fpnameExtendCall},
	}
	err := ast.NewIdent("_err_")
	init := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("_"), err},
		Rhs: []ast.Expr{checkCall},
		Tok: token.DEFINE,
	}

	cond := &ast.BinaryExpr{
		X:  err,
		Op: token.EQL,
		Y:  ast.NewIdent("nil"),
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
	if count := len(call.Args); count > 1 {
		return false, nil, fmt.Errorf("failpoint.Break expect 1 or 0 arguments, but got %v in %s", count, r.pos(call.Pos()))
	}
	var stmt *ast.BranchStmt
	if len(call.Args) > 0 {
		label := call.Args[0].(*ast.BasicLit).Value
		label = strings.Trim(label, "`\"")
		stmt = &ast.BranchStmt{
			TokPos: call.Pos(),
			Tok:    token.BREAK,
			Label:  ast.NewIdent(label),
		}
	} else {
		stmt = &ast.BranchStmt{
			TokPos: call.Pos(),
			Tok:    token.BREAK,
		}
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteContinue(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if count := len(call.Args); count > 1 {
		return false, nil, fmt.Errorf("failpoint.Continue expect 1 or 0 arguments, but got %v in %s", count, r.pos(call.Pos()))
	}
	var stmt *ast.BranchStmt
	if len(call.Args) > 0 {
		label := call.Args[0].(*ast.BasicLit).Value
		label = strings.Trim(label, "`\"")
		stmt = &ast.BranchStmt{
			TokPos: call.Pos(),
			Tok:    token.CONTINUE,
			Label:  ast.NewIdent(label),
		}
	} else {
		stmt = &ast.BranchStmt{
			TokPos: call.Pos(),
			Tok:    token.CONTINUE,
		}
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteLabel(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if count := len(call.Args); count != 1 {
		return false, nil, fmt.Errorf("failpoint.Label expect 1 arguments, but got %v in %s", count, r.pos(call.Pos()))
	}
	label := call.Args[0].(*ast.BasicLit).Value
	label = strings.Trim(label, "`\"")
	stmt := &ast.LabeledStmt{
		Colon: call.Pos(),
		Label: ast.NewIdent(label + labelSuffix), // It's a trick here
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteGoto(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if count := len(call.Args); count != 1 {
		return false, nil, fmt.Errorf("failpoint.Goto expect 1 arguments, but got %v in %s", count, r.pos(call.Pos()))
	}
	label := call.Args[0].(*ast.BasicLit).Value
	label = strings.Trim(label, "`\"")
	stmt := &ast.BranchStmt{
		TokPos: call.Pos(),
		Tok:    token.GOTO,
		Label:  ast.NewIdent(label),
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteFallthrough(call *ast.CallExpr) (bool, ast.Stmt, error) {
	stmt := &ast.BranchStmt{
		TokPos: call.Pos(),
		Tok:    token.FALLTHROUGH,
	}
	return true, stmt, nil
}

func (r *Rewriter) rewriteReturn(call *ast.CallExpr) (bool, ast.Stmt, error) {
	stmt := &ast.ReturnStmt{
		Return:  call.Pos(),
		Results: call.Args,
	}
	return true, stmt, nil
}
