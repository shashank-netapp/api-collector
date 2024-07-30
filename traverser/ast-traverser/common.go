package ast_traverser

import (
	"go/ast"
	"go/token"
)

type Visitor struct {
	LastVisitedNode ast.Node
	Fset            *token.FileSet
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch typeNode := node.(type) {
	case *ast.FuncDecl:
		funcPos := v.fset.Position(typeNode.Name.Pos())
		line := funcPos.Line
		character := funcPos.Column
		fileName := funcPos.Filename
		r := recurseTraversal{
			pack: v.pack,
			fset: v.fset,
		}
		//wg.Add(1)
		// Indexing starts from 1, hence minus 1.
		r.traverseRecursively(fileName, line-1, character-1, typeNode.Name.Name, 0)
	}

	resVisitor := Visitor{
		fset:            v.fset,
		lastVisitedNode: node,
		pack:            v.pack,
	}

	return &resVisitor
}
