package recurser

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"strings"
	"sync"

	"github.com/theshashankpal/api-collector/callgraph"
	"github.com/theshashankpal/api-collector/loader"
	. "github.com/theshashankpal/api-collector/logger"
)

var rr = LogFields{Key: "layer", Value: "rest-dfs-recurser"}

type RESTRecurser struct {
	fset *token.FileSet    // Can fset var be shared ?
	pkgs []*loader.Package // Can package var be shared?

	// Callgraph can be shared between rest and zapi recurser
	callGraph   callgraph.CallGraph
	callgraphMU *sync.Mutex

	visited      map[string]struct{}
	visitedMutex *sync.Mutex

	restAPIs      map[string][]string
	restAPIsMutex *sync.Mutex

	// Don't need a mutex for fileMap, as it is read-only
	fileMap map[string]*ast.File
	wg      *sync.WaitGroup
}

func NewRESTRecurser(callGraph callgraph.CallGraph, callGraphMU *sync.Mutex) *RESTRecurser {
	return &RESTRecurser{
		callGraph:     callGraph,
		callgraphMU:   callGraphMU,
		visited:       make(map[string]struct{}),
		visitedMutex:  new(sync.Mutex),
		restAPIs:      make(map[string][]string),
		restAPIsMutex: new(sync.Mutex),
		wg:            new(sync.WaitGroup),
	}
}

func (r *RESTRecurser) Traverse(ctx context.Context, restAPIsMapChan chan map[string][]string) {
	go func() {
		for _, pkg := range r.pkgs {
			if strings.Contains(pkg.PkgPath, "github.com/netapp/trident/storage_drivers/ontap/api") {
				for _, file := range pkg.Syntax {
					filePath := pkg.Fset.File(file.Package).Name()
					if strings.Contains(filePath, "storage_drivers/ontap/api/ontap_rest.go") {
						fileSet := pkg.Fset
						ast.Inspect(file, func(node ast.Node) bool {
							switch typeNode := node.(type) {
							case *ast.FuncDecl:
								funcPos := fileSet.Position(typeNode.Name.Pos())
								line := funcPos.Line
								character := funcPos.Column
								filePath = funcPos.Filename
								r.fset = fileSet
								r.wg.Add(1)
								//Indexing starts from 1, hence minus 1.
								go r.traverseRecursively(ctx, filePath, line-1, character-1, typeNode.Name.Name)
								return false
							}
							return true
						})
						break
					}
				}
			}
		}
		r.wg.Wait()
		restAPIsMapChan <- r.restAPIs
	}()
}

/*
Why with depth can't use a visited map:
Take example of JobGet function:
First time we reach jobGet through another function with depth 2
then we cannot explore its callees as they will be depth 3, and we're returning in depth 3.
And afterward, when we actually reach jobGet with depth 0, it has been already visited.
*/
func (r *RESTRecurser) traverseRecursively(ctx context.Context, filePath string, line, character int, functionName string) {
	defer r.wg.Done()

	functionID := fmt.Sprintf("%s:%d:%d:%s", filePath, line, character, functionName)

	r.visitedMutex.Lock()
	if _, ok := r.visited[functionID]; ok {
		r.visitedMutex.Unlock()
		return
	}

	r.visited[functionID] = struct{}{}
	r.visitedMutex.Unlock()

	Log(ctx, rr).Trace().
		Str("functionName", functionName).
		Str("functionID", functionID).
		Msg("Visiting function")

	// There can be a situation that we are in some prefix client.go file, it has interface but don't have
	// the corresponding implementation of it in the same file.Then we need to continue and go to the file
	// where the actual implementation is.
	// Also, functionID will be different in that case, as filePath will be different, so if the above is the case,
	// we'll be visiting it and not returning early.
	if strings.Contains(filePath, "ontap/api/rest/client") && strings.HasSuffix(filePath, "client.go") {
		lis := r.restScraper(ctx, filePath, functionName)
		// We've found what we were looking for, so we can return
		// otherwise continue with finding its callees.
		if len(lis) != 0 {
			r.restAPIsMutex.Lock()
			if _, ok := r.restAPIs[functionID]; !ok {
				r.restAPIs[functionID] = lis
			}
			r.restAPIsMutex.Unlock()
			return
		}
	}

	r.callgraphMU.Lock()
	outgoingCallsChan := r.callGraph.OutgoingCalls(ctx, filePath, line, character)
	r.callgraphMU.Unlock()
	outgoingCalls := <-outgoingCallsChan
	if outgoingCalls.Error != nil {
		Log(ctx).Error().
			Int("ErrorCode", outgoingCalls.Error.Code).
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
		isInterface := false

		// Get the file
		file, ok := r.fileMap[filePath]
		if !ok {
			Log(ctx, rr).Panic().Str("filePath", filePath).Stack().Msg("File not found in fileMap")
		}

		var lastTypeSpec *ast.TypeSpec
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.TypeSpec:
				lastTypeSpec = node // Update lastTypeSpec with the current *ast.TypeSpec node
			case *ast.InterfaceType:
				if lastTypeSpec != nil {
					var interfaceType *ast.InterfaceType
					if interfaceType, ok = lastTypeSpec.Type.(*ast.InterfaceType); !ok {
						return true
					}

					for _, method := range interfaceType.Methods.List {
						// method.Names represent: field/method/(type) parameter names; or nil
						// here it is just a method, so Names[0] should do
						if len(method.Names) > 0 {
							interfacName := method.Names[0].Name
							linePos := method.Names[0].Pos()
							lineInteface := r.fset.Position(linePos).Line
							if interfacName == functionName && line == (lineInteface-1) {
								isInterface = true
								return false
							}
						}
					}
				}
			}
			return true
		})

		if isInterface == true {
			r.callgraphMU.Lock()
			implementationsChan := r.callGraph.Implementations(ctx, filePath, line, character)
			r.callgraphMU.Unlock()
			implementation := <-implementationsChan
			if implementation.Error != nil {
				Log(ctx, rr).Error().
					Int("ErrorCode", outgoingCalls.Error.Code).
					Str("Error", outgoingCalls.Error.Message).
					Str("FilePath", filePath).
					Str("FunctionName", functionName).
					Int("Character", character).
					Int("Line", line).
					Msg("Error getting implementation")
				return
			}
			for _, impl := range implementation.Result {
				if strings.Contains(impl.Uri, "mocks") {
					continue
				}
				filePath = impl.Uri
				filePath = strings.ReplaceAll(filePath, "file://", "")
				r.wg.Add(1)
				go r.traverseRecursively(ctx, filePath, impl.Range.Start.Line, impl.Range.Start.Character, functionName)
			}
		}
	}

	for _, call := range outgoingCalls.Result {
		// Don't want to explore callee of other packages
		if !strings.Contains(call.To.Detail, "github.com/netapp/trident") {
			continue
		}

		if !strings.Contains(call.To.Detail, "github.com/netapp/trident/storage_drivers/ontap/api") {
			continue
		}

		line = call.To.Range.Start.Line
		character = call.To.Range.Start.Character
		filePath = call.To.Uri
		filePath = strings.ReplaceAll(filePath, "file://", "")
		r.wg.Add(1)
		go r.traverseRecursively(ctx, filePath, line, character, call.To.Name)
	}
}

func (r *RESTRecurser) restScraper(ctx context.Context, filePath, functionName string) []string {
	Log(ctx, rr).Debug().Str("filePath", filePath).Str("functionName", functionName).Msg("Found REST API")

	file, ok := r.fileMap[filePath]
	if !ok {
		Log(ctx, rr).Panic().Str("filePath", filePath).Stack().Msg("File not found in fileMap")
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

	if method != "" && api != "" {
		Log(ctx, rr).Debug().Str("Method", method).Str("API", api).Msg("REST API")
		lis := []string{method, api}
		return lis
	}

	return nil
}

func (r *RESTRecurser) SetFileSet(fset *token.FileSet) {
	r.fset = fset
}

func (r *RESTRecurser) SetFileMap(fileMap map[string]*ast.File) {
	r.fileMap = fileMap
}

func (r *RESTRecurser) SetPackages(pkgs []*loader.Package) {
	r.pkgs = pkgs
}
