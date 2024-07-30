package dfs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/theshashankpal/api-collector/callgraph"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"strings"

	"github.com/theshashankpal/api-collector/callgraph/lsp/requests"
	"github.com/theshashankpal/api-collector/loader"
	. "github.com/theshashankpal/api-collector/logger"
	. "github.com/theshashankpal/api-collector/traverser/ast-traverser"
)

var dfsF = LogFields{Key: "layer", Value: "dfs-traverser"}

type DfsTraverser struct {
	pack []*loader.Package
}

func NewDfsTraverser(pack []*loader.Package) *DfsTraverser {
	return &DfsTraverser{
		pack: pack,
	}
}

func (d *DfsTraverser) traverseREST(ctx context.Context, recurser recurserInterface) {
	for _, pkg := range d.pack {
		if strings.Contains(pkg.PkgPath, "github.com/netapp/trident/storage_drivers/ontap/api") {
			for _, file := range pkg.Syntax {
				filePath := pkg.Fset.File(file.Package).Name()
				if strings.Contains(filePath, "trident/storage_drivers/ontap/api/ontap_rest.go") {
					//visitorStruct := &Visitor{
					//	Fset: pkg.Fset,
					//}
					//ast.Walk(visitorStruct, file)
					fileSet := pkg.Fset
					ast.Inspect(file, func(node ast.Node) bool {
						switch typeNode := node.(type) {
						case *ast.FuncDecl:
							funcPos := fileSet.Position(typeNode.Name.Pos())
							line := funcPos.Line
							character := funcPos.Column
							fileName := funcPos.Filename
							r := recurseTraversal{
								fset: fileSet,
							}
							//wg.Add(1)
							// Indexing starts from 1, hence minus 1.
							r.traverseRecursively(fileName, line-1, character-1, typeNode.Name.Name, 0)
						}

						return false
					})
					break
				}
			}
		}
	}
}

func (d *DfsTraverser) traverseZAPI() {

}


func (r *recurseTraversal) traverserZAPI(filePath string, line, character int, functionName string, depth int) {
	//defer wg.Done()
	if depth == 3 {
		return
	}

	functionID := fmt.Sprintf("%s:%d:%d:%s", filePath, line, character, functionName)

	//mutex.Lock()
	//if _, ok := visited[functionID]; ok {
	//	//	//mutex.Unlock()
	//	return
	//}

	/*
		Why with depth can't use a visited map:
		Take example of JobGet function:
		First time we reach jobGet through another function with depth 2
		then we cannot explore its callees as they will be depth 3, and we're returning in depth 3.
		And afterward, when we actually reach jobGet with depth 0, it has been already visited.
	*/
	//visited[functionID] = struct{}{}

	// There can be a situation that we are in some prefix client.go file, it has interface but don't have
	// the corresponding implementation of it in the same file.Then we need to continue and go to the file
	// where the actual implementation is.
	// Also, functionID will be different in that case, as filePath will be different, so if the above is the case,
	// we'll be visiting it and not returning early.
	if strings.Contains(filePath, "ontap/api/rest/client") && strings.HasSuffix(filePath, "client.go") {
		fmt.Println("FileName: ", filePath)
		fmt.Println("Function Name: ", functionName)
		file, ok := fileMap[filePath]
		if !ok {
			panic("File not found")
			return
		}
		var method string
		var api string

		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true // not a FuncDecl, skip this node
			}

			// Check if the function has the name we're looking for
			if funcDecl.Name.Name != functionName {
				return true // not the function we're looking for
			}

			var buf bytes.Buffer

			// Reading the function body
			err := printer.Fprint(&buf, r.fset, funcDecl)
			if err != nil {
				panic(err)
			}

			// Creating a reader to read the function body.
			reader := bufio.NewReader(&buf)
			for {
				// Reading line by line.
				line, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}

				if err != nil {
					panic(err)
				}

				// Checking if the line contains the Method
				// Example:
				//	Method:             "POST",
				if strings.Contains(line, "Method") {
					parts := strings.Split(line, ":")         // [{`Method`,`"POST",`}
					method = strings.TrimSpace(parts[1])      // "POST",
					method = strings.TrimSuffix(method, `",`) // "POST
					method = strings.TrimPrefix(method, `"`)  // POST
				}

				// 	Checking if the line contains the PathPattern
				// 	Example:
				//	PathPattern:        "/storage/volumes/{volume.uuid}/snapshots",
				if strings.Contains(line, "PathPattern") {
					parts := strings.Split(line, ":")   // [{`PathPattern`,`"/storage/volumes/{volume.uuid}/snapshots",`}
					api = strings.TrimSpace(parts[1])   // "/storage/volumes/{volume.uuid}/snapshots",
					api = strings.TrimSuffix(api, `",`) // "/storage/volumes/{volume.uuid}/snapshots
					api = strings.TrimPrefix(api, `"`)  // /storage/volumes/{volume.uuid}/snapshots
					return false                        // stop the AST traversal
				}
			}
			return true
		})

		fmt.Println("Method: ", method)
		fmt.Println("API: ", api)
		lis := []string{method, api}

		// We've found what we were looking for, so we can return
		// otherwise continue with finding its callees.
		if method != "" && api != "" {
			if _, ok := apiMap[functionID]; !ok {
				apiMap[functionID] = lis
			}
			return
		}
	}

	//mutex.Unlock()
	//fmt.Printf("Visiting function: %s in file: %s\n", functionName, filePath)

	outgoingCalls, err := findCallees(filePath, line, character)
	if err != nil {
		fmt.Println("Error finding callees: ", err)
		fmt.Println("FileName: ", filePath)
		fmt.Println("Function Name: ", functionName)
		fmt.Println("Line: ", line)
		fmt.Println("Character: ", character)
	}

	// outGoingCalls can be of length 0, indicating that we might have encountered an interface.
	if len(outgoingCalls) == 0 {
		documentSymbols, _ := getAllTheSymbols(filePath)
		isInterface := false
		for _, documentSymbol := range documentSymbols {
			if documentSymbol.Kind == requests.SymbolKindInterface {
				if documentSymbol.Location.Range.Start.Line < line && documentSymbol.Location.Range.End.Line > line {
					isInterface = true
					break
				}
			}
		}

		if isInterface == true {
			implementations, err := findImplementation(filePath, line, character)
			for _, implementation := range implementations {
				if strings.Contains(implementation.Uri, "mocks") {
					continue
				}
				filePath = implementation.Uri
				filePath = strings.ReplaceAll(filePath, "file://", "")
				//wg.Add(1)
				r.traverseRecursively(filePath, implementation.Range.Start.Line, implementation.Range.Start.Character, functionName, depth+1)
			}
			if err != nil {
				fmt.Println("Error getting implementations: ", err)
			}
		}
	}

	for _, call := range outgoingCalls {
		// Don't want to explore callee of other packages
		if !strings.Contains(call.To.Detail, "github.com/netapp/trident") {
			continue
		}

		if strings.Contains(call.To.Detail, "github.com/netapp/trident/logging") {
			continue
		}

		line = call.To.Range.Start.Line
		character = call.To.Range.Start.Character
		filePath = call.To.Uri
		filePath = strings.ReplaceAll(filePath, "file://", "")
		//wg.Add(1)
		r.traverseRecursively(filePath, line, character, call.To.Name, depth+1)
	}

	return

}
