package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

var (
	treLabel = &ast.Ident{
		Name: "__tre__",
	}
)

type visitor func(ast.Node) bool

func (v visitor) Visit(node ast.Node) ast.Visitor {
	if v(node) {
		return v
	} else {
		return nil
	}
}

func isTailCall(ret *ast.ReturnStmt, f *ast.FuncDecl) bool {
	if len(ret.Results) != 1 {
		return false
	}
	if call, ok := ret.Results[0].(*ast.CallExpr); ok {
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if ident.Name == f.Name.Name {
				return true
			}
		}
	}
	return false
}

func fieldListToExpr(fields *ast.FieldList) []ast.Expr {
	result := make([]ast.Expr, 0)
	for _, field := range fields.List {
		for _, name := range field.Names {
			result = append(result, name)
		}
	}
	return result
}

func tailRecursionOptimize(f *ast.FuncDecl) *ast.FuncDecl {
	trePossible := false
	ast.Walk(visitor(func(node ast.Node) bool {
		if ret, ok := node.(*ast.ReturnStmt); ok {
			trePossible = isTailCall(ret, f)
		}
		return !trePossible
	}), f)

	if !trePossible {
		return f
	}

	f.Body.List[0] = &ast.LabeledStmt{
		Label: treLabel,
		Stmt:  f.Body.List[0],
	}

	ast.Walk(visitor(func(node ast.Node) bool {
		if block, ok := node.(*ast.BlockStmt); ok {
			if len(block.List) == 0 {
				return true
			}

			if ret, ok := block.List[len(block.List)-1].(*ast.ReturnStmt); ok && isTailCall(ret, f) {
				block.List[len(block.List)-1] = &ast.AssignStmt{
					Lhs: fieldListToExpr(f.Type.Params),
					Tok: token.ASSIGN,
					Rhs: ret.Results[0].(*ast.CallExpr).Args,
				}

				block.List = append(block.List,
					&ast.BranchStmt{
						Tok:   token.GOTO,
						Label: treLabel,
					},
				)
			}
		}
		return true
	}), f)
	return f
}

func main() {
	if len(os.Args) != 2 {
		panic("Please give file to optimize")
	}

	fset := token.NewFileSet()
	inputFile, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for i, decl := range inputFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Recv == nil {
			inputFile.Decls[i] = tailRecursionOptimize(funcDecl)
		}
	}

	format.Node(os.Stdout, fset, inputFile)
}
