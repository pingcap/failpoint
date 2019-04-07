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
	"strings"
)

type exprRewriter func(rewriter *Rewriter, call *ast.CallExpr) (rewritten bool, result ast.Stmt, err error)

var exprRewriters = map[string]exprRewriter{
	"Inject":      (*Rewriter).rewriteInject,
	"Break":       (*Rewriter).rewriteBreak,
	"Continue":    (*Rewriter).rewriteContinue,
	"Label":       (*Rewriter).rewriteLabel,
	"Goto":        (*Rewriter).rewroteGoto,
	"Fallthrough": (*Rewriter).rewroteFallthrough,
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

	if len(fpbody.Type.Params.List) > 2 {
		var types []string
		for _, field := range fpbody.Type.Params.List {
			types = append(types, fmt.Sprintf("%T", field.Type))
		}
		return false, nil, fmt.Errorf("failpoint: invalid signature(%s)", strings.Join(types, ", "))
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
		var ctx, arg *ast.Field
		for _, field := range fpbody.Type.Params.List {
			selector, ok := field.Type.(*ast.SelectorExpr)
			if !ok {
				return false, nil, fmt.Errorf("failpoint: invalid signature(%T, %T)",
					fpbody.Type.Params.List[0].Type, fpbody.Type.Params.List[1].Type)
			}
			switch {
			case selector.Sel.Name == "Context":
				ctx = field
			case selector.Sel.Name == "Value" && selector.X.(*ast.Ident).Name == r.failpointName:
				arg = field
			default:
				return false, nil, fmt.Errorf("failpoint: invalid signature with type: %T", field.Type)
			}
		}
		if ctx == nil || arg == nil {
			return false, nil, fmt.Errorf("failpoint: invalid signature(%T, %T)",
				fpbody.Type.Params.List[0].Type, fpbody.Type.Params.List[1].Type)
		}
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
		selector, ok := arg.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Value" || selector.X.(*ast.Ident).Name != r.failpointName {
			return false, nil, fmt.Errorf("failpoint: invalid signature with type: %T", arg.Type)
		}

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
	if count := len(call.Args); count > 1 {
		return false, nil, fmt.Errorf("failpoint.Break expect 1 or 0 arguments, but got %v", count)
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
		return false, nil, fmt.Errorf("failpoint.Continue expect 1 or 0 arguments, but got %v", count)
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
		return false, nil, fmt.Errorf("failpoint.Label expect 1 arguments, but got %v", count)
	}
	label := call.Args[0].(*ast.BasicLit).Value
	label = strings.Trim(label, "`\"")
	stmt := &ast.LabeledStmt{
		Colon: call.Pos(),
		Label: ast.NewIdent(label),
	}
	return true, stmt, nil
}

func (r *Rewriter) rewroteGoto(call *ast.CallExpr) (bool, ast.Stmt, error) {
	if count := len(call.Args); count != 1 {
		return false, nil, fmt.Errorf("failpoint.Goto expect 1 arguments, but got %v", count)
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

func (r *Rewriter) rewroteFallthrough(call *ast.CallExpr) (bool, ast.Stmt, error) {
	stmt := &ast.BranchStmt{
		TokPos: call.Pos(),
		Tok:    token.FALLTHROUGH,
	}
	return true, stmt, nil
}
