package recurser

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"sync"

	"github.com/theshashankpal/api-collector/callgraph"
	"github.com/theshashankpal/api-collector/loader"
	. "github.com/theshashankpal/api-collector/logger"
)

var zr = LogFields{Key: "layer", Value: "zapi-dfs-recurser"}

type ZAPIRecurser struct {
	fset *token.FileSet    // Can fset var be shared ?
	pkgs []*loader.Package // Can package var be shared?

	// Callgraph can be shared between rest and zapi recurser
	callGraph   callgraph.CallGraph
	callgraphMU *sync.Mutex

	visited      map[string]struct{}
	visitedMutex *sync.Mutex

	zapiCMDs   map[string][]string
	zapiCMDsMU *sync.Mutex

	// Don't need a mutex for fileMap, as it is read-only
	fileMap map[string]*ast.File
	wg      *sync.WaitGroup
}

func NewZAPIRecurser(callGraph callgraph.CallGraph, callGraphMU *sync.Mutex) *ZAPIRecurser {
	return &ZAPIRecurser{
		callGraph:    callGraph,
		callgraphMU:  callGraphMU,
		visited:      make(map[string]struct{}),
		visitedMutex: new(sync.Mutex),
		zapiCMDs:     make(map[string][]string),
		zapiCMDsMU:   new(sync.Mutex),
		wg:           new(sync.WaitGroup),
	}
}

func (z *ZAPIRecurser) Traverse(ctx context.Context, zapiCommandsChan chan map[string][]string) {
	go func() {
		for _, pkg := range z.pkgs {
			if strings.Contains(pkg.PkgPath, "github.com/netapp/trident/storage_drivers/ontap/api") {
				for _, file := range pkg.Syntax {
					filePath := pkg.Fset.File(file.Package).Name()
					if strings.Contains(filePath, "storage_drivers/ontap/api/ontap_zapi.go") {
						fileSet := pkg.Fset
						ast.Inspect(file, func(node ast.Node) bool {
							switch typeNode := node.(type) {
							case *ast.FuncDecl:
								funcPos := fileSet.Position(typeNode.Name.Pos())
								line := funcPos.Line
								character := funcPos.Column
								filePath = funcPos.Filename
								z.fset = fileSet
								z.wg.Add(1)
								//Indexing starts from 1, hence minus 1.
								go z.traverseRecursively(ctx, filePath, line-1, character-1, typeNode.Name.Name)
								return false
							}
							return true
						})
						break
					}
				}
			}
		}
		z.wg.Wait()
		zapiCommandsChan <- z.zapiCMDs
	}()
}

/*
Why with depth can't use a visited map:
Take example of JobGet function:
First time we reach jobGet through another function with depth 2
then we cannot explore its callees as they will be depth 3, and we're returning in depth 3.
And afterward, when we actually reach jobGet with depth 0, it has been already visited.
*/
func (z *ZAPIRecurser) traverseRecursively(ctx context.Context, filePath string, line, character int, functionName string) {
	defer z.wg.Done()

	functionID := fmt.Sprintf("%s:%d:%d:%s", filePath, line, character, functionName)

	z.visitedMutex.Lock()
	if _, ok := z.visited[functionID]; ok {
		z.visitedMutex.Unlock()
		return
	}

	z.visited[functionID] = struct{}{}
	z.visitedMutex.Unlock()

	Log(ctx, zr).Trace().
		Str("functionName", functionName).
		Str("functionID", functionID).
		Msg("Visiting function")

	// There can be a situation that we are in some prefix client.go file, it has interface but don't have
	// the corresponding implementation of it in the same file.Then we need to continue and go to the file
	// where the actual implementation is.
	// Also, functionID will be different in that case, as filePath will be different, so if the above is the case,
	// we'll be visiting it and not returning early.

	if strings.Contains(filePath, "ontap/api/azgo") {
		pathParts := strings.Split(filePath, "/")
		fileEnd := pathParts[len(pathParts)-1]
		if strings.HasPrefix(fileEnd, "api-") {
			if functionName == "ExecuteUsing" {
				command := z.zapiScraper(ctx, filePath, functionName)
				// We've found what we were looking for, so we can return
				// otherwise continue with finding its callees.
				if len(command) != 0 {
					z.zapiCMDsMU.Lock()
					if _, ok := z.zapiCMDs[functionID]; !ok {
						z.zapiCMDs[functionID] = command
					}
					z.zapiCMDsMU.Unlock()
					return
				}
			}
		}
	}

	//mutex.Unlock()
	//fmt.Printf("Visiting function: %s in file: %s\n", functionName, filePath)

	z.callgraphMU.Lock()
	outgoingCallsChan := z.callGraph.OutgoingCalls(ctx, filePath, line, character)
	z.callgraphMU.Unlock()
	outgoingCalls := <-outgoingCallsChan
	if outgoingCalls.Error != nil {
		Log(ctx, zr).Error().
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
		file, ok := z.fileMap[filePath]
		if !ok {
			Log(ctx, zr).Panic().Str("filePath", filePath).Stack().Msg("File not found in fileMap")
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
							lineInteface := z.fset.Position(linePos).Line
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
			z.callgraphMU.Lock()
			implementationsChan := z.callGraph.Implementations(ctx, filePath, line, character)
			z.callgraphMU.Unlock()
			implementation := <-implementationsChan
			if implementation.Error != nil {
				Log(ctx).Error().
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
				z.wg.Add(1)
				go z.traverseRecursively(ctx, filePath, impl.Range.Start.Line, impl.Range.Start.Character, functionName)
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
		z.wg.Add(1)
		go z.traverseRecursively(ctx, filePath, line, character, call.To.Name)
	}
}

func (z *ZAPIRecurser) zapiScraper(ctx context.Context, filePath string, functionName string) []string {
	Log(ctx, zr).Debug().Str("filePath", filePath).Str("functionName", functionName).Msg("Found ZAPI Command")

	file, ok := z.fileMap[filePath]
	if !ok {
		Log(ctx, zr).Panic().Str("filePath", filePath).Stack().Msg("File not found in fileMap")
	}
	var command = make([]string, 1)
	ast.Inspect(file, func(node ast.Node) bool {
		typeNode, ok := node.(*ast.FuncDecl)
		if !ok {
			return true // not a FuncDecl, skip this node
		}

		if typeNode.Name.Name != functionName {
			return true // not the function we're looking for
		}

		reciverType := typeNode.Recv.List[0].Type
		var receiverStarExpr *ast.StarExpr
		if receiverStarExpr, ok = reciverType.(*ast.StarExpr); !ok {
			Log(ctx, zr).Panic().Stack().Msgf("Receiver is not a star expression")
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
				Log(ctx, zr).Debug().Str("ZAPI Command", value).Msg("ZAPI Command")
				command[0] = value
				break
			}
		}
		return false
	})

	return command
}

func (z *ZAPIRecurser) SetFileSet(fset *token.FileSet) {
	z.fset = fset
}

func (z *ZAPIRecurser) SetFileMap(fileMap map[string]*ast.File) {
	z.fileMap = fileMap
}

func (z *ZAPIRecurser) SetPackages(pkgs []*loader.Package) {
	z.pkgs = pkgs
}
