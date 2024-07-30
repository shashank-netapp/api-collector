package dfs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/theshashankpal/api-collector/callgraph/lsp/requests"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"strings"
	"sync"

	"github.com/theshashankpal/api-collector/callgraph"
	. "github.com/theshashankpal/api-collector/logger"
)

var rec = LogFields{Key: "layer", Value: "dfs-recurser"}

type recurserInterface interface {
}

type Recurser struct {
	fset              *token.FileSet
	rest              bool
	zapi              bool
	lsp               callgraph.CallGraph
	visited           map[string]struct{}
	visitedMutex      *sync.Mutex
	restAPIs          map[string][]string
	restAPIsMutex     *sync.Mutex
	zapiCommands      map[string][]string
	zapiCommandsMutex *sync.Mutex
	fileMap           map[string]*ast.File
}

func NewRecurser(fs *token.FileSet, rest bool, zapi bool, lsp callgraph.CallGraph) *Recurser {
	return &Recurser{
		fset:              fs,
		rest:              rest,
		zapi:              zapi,
		lsp:               lsp,
		visited:           make(map[string]struct{}),
		restAPIs:          make(map[string][]string),
		zapiCommands:      make(map[string][]string),
		visitedMutex:      new(sync.Mutex),
		restAPIsMutex:     new(sync.Mutex),
		zapiCommandsMutex: new(sync.Mutex),
	}
}

/*
Why with depth can't use a visited map:
Take example of JobGet function:
First time we reach jobGet through another function with depth 2
then we cannot explore its callees as they will be depth 3, and we're returning in depth 3.
And afterward, when we actually reach jobGet with depth 0, it has been already visited.
*/
func (r *Recurser) traverseRecursively(ctx context.Context, filePath string, line, character int, functionName string) {
	//defer wg.Done()

	functionID := fmt.Sprintf("%s:%d:%d:%s", filePath, line, character, functionName)

	r.visitedMutex.Lock()
	if _, ok := r.visited[functionID]; ok {
		r.visitedMutex.Unlock()
		return
	}

	Log(ctx, rec).Debug().
		Str("functionName", functionName).
		Str("functionID", functionID).
		Msg("Visiting function")

	r.visited[functionID] = struct{}{}
	r.visitedMutex.Unlock()

	// There can be a situation that we are in some prefix client.go file, it has interface but don't have
	// the corresponding implementation of it in the same file.Then we need to continue and go to the file
	// where the actual implementation is.
	// Also, functionID will be different in that case, as filePath will be different, so if the above is the case,
	// we'll be visiting it and not returning early.
	if r.rest {
		lis := r.restScraper(ctx, filePath, functionName)
		if len(lis) != 0 {
			r.restAPIsMutex.Lock()
			if _, ok := r.restAPIs[functionID]; !ok {
				r.restAPIs[functionID] = lis
			}
			r.restAPIsMutex.Unlock()
			return
		}
	}

	//mutex.Unlock()
	//fmt.Printf("Visiting function: %s in file: %s\n", functionName, filePath)

	outgoingCallsChan := r.lsp.OutgoingCalls(ctx, filePath, line, character)
	outgoingCalls := <-outgoingCallsChan
	if outgoingCalls.Error != nil {
		Log(ctx).Error().
			Str("Error", outgoingCalls.Error.Message).
			Str("FilePath", filePath).
			Str("FunctionName", functionName).
			Int("Character", character).
			Int("Line", line).
			Msg("Error getting outgoing calls")
		return
	}

	// outGoingCalls can be of length 0, indicating that we might have encountered an interface.
	if len(outgoingCalls.Result) == 0 {
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

	for _, call := range outgoingCalls.Result {
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

func (r *Recurser) zapiScraper(filePath string, line, character int, functionName string) {

	// For ZAPI
	if strings.Contains(filePath, "ontap/api/azgo") {
		pathParts := strings.Split(filePath, "/")
		fileEnd := pathParts[len(pathParts)-1]
		if strings.HasPrefix(fileEnd, "api-") {
			if functionName == "ExecuteUsing" {
				file, ok := fileMap[filePath]
				if !ok {
					panic("File not found")
					return
				}
				ast.Inspect(file, func(node ast.Node) bool {
					typeNode, ok := node.(*ast.FuncDecl)
					if !ok {
						return true // not a FuncDecl, skip this node
					}

					if typeNode.Name.Name != functionName {
						return true // not the function we're looking for
					}

					fmt.Println("Found")

					reciverType := typeNode.Recv.List[0].Type
					var receiverStarExpr *ast.StarExpr
					if receiverStarExpr, ok = reciverType.(*ast.StarExpr); !ok {
						panic("Receiver is not a star expression")
					}
					receiverIdent := receiverStarExpr.X.(*ast.Ident)
					receiverStruct := receiverIdent.Obj.Decl.(*ast.TypeSpec).Type.(*ast.StructType)
					for _, field := range receiverStruct.Fields.List {
						if field.Type.(*ast.SelectorExpr).X.(*ast.Ident).Name == "xml" &&
							field.Type.(*ast.SelectorExpr).Sel.Name == "Name" {
							tag := field.Tag
							tagParts := strings.Split(tag.Value, ":")
							value := tagParts[1]
							value = strings.TrimSuffix(value, "\"`")
							value = strings.TrimPrefix(value, "\"")
							fmt.Println("Command: ", value)
							commandMap[functionID] = value
							return false
						}
					}
					return false
				})
			}
		}
	}
}

func (r *Recurser) restScraper(ctx context.Context, filePath, functionName string) []string {
	if strings.Contains(filePath, "ontap/api/rest/client") && strings.HasSuffix(filePath, "client.go") {
		Log(ctx, rec).Debug().Str("filePath", filePath).Str("functionName", functionName).Msg("Found REST API")

		file, ok := r.fileMap[filePath]
		if !ok {
			Log(ctx, rec).Panic().Str("filePath", filePath).Stack().Msg("File not found in fileMap")
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

		// We've found what we were looking for, so we can return
		// otherwise continue with finding its callees.
		if method != "" && api != "" {
			Log(ctx, rec).Debug().Str("Method", method).Str("API", api).Msg("REST API")
			lis := []string{method, api}
			return lis
		}
	}

	return nil
}

func (r *Recurser) SetFileSet(fileset *token.FileSet) {
	r.fset = fileset
}
