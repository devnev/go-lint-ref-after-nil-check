package nilref

import (
	"go/ast"
	"go/token"
)

func Check(f *ast.File) []ast.Stmt {
	var c checkVisitor
	ast.Walk(&c, f)
	return c.failures
}

type checkVisitor struct {
	failures []ast.Stmt
}

func (v *checkVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}
	block, ok := node.(*ast.BlockStmt)
	if !ok {
		return v
	}
	for idx, stmt := range block.List {
		cond, ok := stmt.(*ast.IfStmt)
		if !ok {
			continue
		}
		errObj := isNilCheck(cond)
		if errObj == nil {
			continue
		}
		var remainder []ast.Stmt
		if cond.Else != nil {
			remainder = append(remainder, cond.Else)
		}
		remainder = append(remainder, block.List[idx+1:]...)
		if stmt := checkForRefs(errObj, remainder); stmt != nil {
			v.failures = append(v.failures, stmt)
		}
	}
	return v
}

func isNilCheck(ifStmt *ast.IfStmt) *ast.Object {
	if !isTerminal(ifStmt.Body.List) {
		return nil
	}

	bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return nil
	}
	if bin.Op != token.NEQ {
		return nil
	}
	var varExpr, nilExpr *ast.Ident
	if varExpr, ok = bin.X.(*ast.Ident); !ok {
		return nil
	}
	if nilExpr, ok = bin.Y.(*ast.Ident); !ok {
		return nil
	}
	if varExpr.Name == "nil" {
		nilExpr, varExpr = varExpr, nilExpr
	}
	if nilExpr.Obj != nil {
		panic("nil expr " + nilExpr.Name + " has object")
	}
	if varExpr.Obj == nil {
		panic("var expr " + varExpr.Name + " has nil obj")
	}
	return varExpr.Obj
}

// return true if control flow never reaches the end of the given set of statements
func isTerminal(block []ast.Stmt) bool {
	if len(block) == 0 {
		return false
	}
	_, hasReturn := block[len(block)-1].(*ast.ReturnStmt)
	return hasReturn
}

func checkForRefs(obj *ast.Object, block []ast.Stmt) ast.Stmt {
	for _, s := range block {
		written := writeVisitor{
			tgt: obj,
		}
		ast.Walk(&written, s)
		if written.written {
			return nil
		}
		referenced := refVisitor{
			tgt: obj,
		}
		ast.Walk(&referenced, s)
		if referenced.referenced {
			return s
		}
	}
	return nil
}

type writeVisitor struct {
	tgt     *ast.Object
	written bool
}

func (v *writeVisitor) Visit(node ast.Node) ast.Visitor {
	if v.written {
		return nil
	}
	switch n := node.(type) {
	case *ast.UnaryExpr:
		if n.Op != token.AND {
			break
		}
		id, ok := n.X.(*ast.Ident)
		if !ok {
			break
		}
		if id.Obj == v.tgt {
			v.written = true
		}
	case *ast.AssignStmt:
		for _, e := range n.Lhs {
			id, ok := e.(*ast.Ident)
			if !ok {
				continue
			}
			if id.Obj == v.tgt {
				v.written = true
				break
			}
		}
	}
	if v.written {
		return nil
	}
	return v
}

type refVisitor struct {
	tgt        *ast.Object
	referenced bool
}

func (v *refVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil || v.referenced {
		return nil
	}
	id, ok := node.(*ast.Ident)
	if !ok {
		return v
	}
	if id.Obj == v.tgt {
		v.referenced = true
	}
	return v
}
