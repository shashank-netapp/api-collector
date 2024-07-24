package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/theshashankpal/api-collector/loader"
)

var dir = "/Users/shashank/Library/CloudStorage/OneDrive-NetAppInc/Documents/Astra/TRID-POLLING-NEW/trident/..."
var core_file = "/Users/shashank/Library/CloudStorage/OneDrive-NetAppInc/Documents/Astra/TRID-POLLING-NEW/trident/core/orchestrator_core.go"
var core_types_file = "/Users/shashank/Library/CloudStorage/OneDrive-NetAppInc/Documents/Astra/TRID-POLLING-NEW/trident/core/types.go"
var visited = make(map[string]struct{})
var wg sync.WaitGroup
var mutex sync.Mutex
var fileMap = make(map[string]*ast.File)
var apiMap map[string][]string
var finalMap = make(map[string]map[string][]string)

//var functionRequiredMap = map[string]struct{}{
//	"addVolumeInitial":             {},
//	"deleteVolume":                 {},
//	"addBackend":                   {},
//	"ImportVolume":                 {},
//	"CreateSnapshot":               {},
//	"deleteSnapshot":               {},
//	"RestoreSnapshot":              {},
//	"ReadSnapshotsForVolume":       {},
//	"GetMirrorTransferTime":        {},
//	"CheckMirrorTransferState":     {},
//	"UpdateMirror":                 {},
//	"reconcileNodeAccessOnBackend": {},
//	"GetReplicationDetails":        {},
//	"ReleaseMirror":                {},
//	"GetMirrorStatus":              {},
//	"PromoteMirror":                {},
//	"ReestablishMirror":            {},
//	"EstablishMirror":              {},
//	"GetVolumeForImport":           {},
//	"PublishVolume":                {},
//	"CloneVolume":                  {},
//	"deleteBackendByBackendUUID":   {},
//	"ResizeVolume":                 {},
//	"Unpublish":                    {},
//}

var functionRequiredMap = make(map[string]struct{})

func traverse(buildFlags []string) error {

	// Load the package.
	pack, err := loader.LoadRoots(dir)
	if err != nil {
		return err
	}

	for _, pkg := range pack {
		// Adding only the package which contains api calls
		pkg.NeedSyntax()
		for _, file := range pkg.Syntax {
			filePath := pkg.Fset.File(file.Package).Name()
			fileMap[filePath] = file
			if filePath == core_types_file {
				readingMethodsOfOrchestratorInterface(file)
			}
		}
	}

	for _, pkg := range pack {
		if strings.Contains(pkg.PkgPath, "github.com/netapp/trident/core") {
			for _, file := range pkg.Syntax {
				if pkg.Fset.File(file.Package).Name() == core_file {
					visitorStruct := &visitor{
						fset: pkg.Fset,
						pack: pack,
					}
					ast.Walk(visitorStruct, file)
					break
				}
			}
		}
	}

	fmt.Println("--------------------------------MAP_API---------------------------------")
	//for key, value := range apiMap {
	//	lent = lent + len(value)
	//	fmt.Println("Function Name: ", key)
	//	fmt.Println("API: ", value)
	//}
	//fmt.Println(len(apiMap))
	//fmt.Println(lent)

	truncateFile("api.txt")
	for keyFunc, mapFunc := range finalMap {
		appendTofile("api.txt", "--------------------"+keyFunc+"--------------------")
		for key, value := range mapFunc {
			key = strings.TrimSpace(strings.Split(key, ":")[3])
			if len(value) == 0 {
				continue
			}

			method := value[0]
			api := value[1]
			appendTofile("api.txt", fmt.Sprintf("%s : %s : %s", key, method, api))
		}
	}
	return nil

}

func appendTofile(fileName, text string) {
	// Open file in append mode, create it if it does not exist, open in write-only mode
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Write text to file
	_, err = file.WriteString(text + "\n") // Adding a newline for each text added
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
}

func truncateFile(filePath string) {
	// Open the file in write-only mode with the O_TRUNC flag to truncate it to zero length
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer file.Close()
}

type recurseTraversal struct {
	pack []*loader.Package
	fset *token.FileSet
}

type visitor struct {
	lastVisitedNode ast.Node
	fset            *token.FileSet
	pack            []*loader.Package
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch typeNode := node.(type) {
	case *ast.FuncDecl:
		if _, ok := functionRequiredMap[typeNode.Name.Name]; ok {
			//if typeNode.Name.Name == "CreateSnapshot" {
			funcPos := v.fset.Position(typeNode.Name.Pos())
			line := funcPos.Line
			character := funcPos.Column
			fileName := funcPos.Filename
			r := recurseTraversal{
				pack: v.pack,
				fset: v.fset,
			}

			fmt.Println("Processing function: ", typeNode.Name.Name)
			apiMap = make(map[string][]string)
			visited = make(map[string]struct{})
			//wg.Add(1)
			// Indexing starts from 1, hence minus 1.
			r.traverseRecursively(fileName, line-1, character-1, typeNode.Name.Name, 0)

			finalMap[typeNode.Name.Name] = apiMap

		}
	}

	resVisitor := visitor{
		fset:            v.fset,
		lastVisitedNode: node,
		pack:            v.pack,
	}

	return &resVisitor
}

func (r *recurseTraversal) traverseRecursively(fileName string, line, character int, functionName string, depth int) {
	//defer wg.Done()
	if depth == 20 {
		return
	}

	functionID := fmt.Sprintf("%s:%d:%d:%s", fileName, line, character, functionName)

	//mutex.Lock()
	if _, ok := visited[functionID]; ok {
		//	//mutex.Unlock()
		return
	}

	/*
		Why with depth can't use a visited map:
		Take example of JobGet function:
		First time we reach jobGet through another function with depth 2
		then we cannot explore its callees as they will be depth 3, and we're returning in depth 3.
		And afterward, when we actually reach jobGet with depth 0, it has been already visited.
	*/
	visited[functionID] = struct{}{}

	// For ZAPI
	if functionName == "ExecuteUsing" && strings.HasSuffix(fileName, "ontap_zapi.go") {

	}

	// For REST

	// There can be a situation that we are in some prefix client.go file, it has interface but don't have
	// the corresponding implementation of it in the same file.Then we need to continue and go to the file
	// where the actual implementation is.
	// Also, functionID will be different in that case,
	// as fileName, line, character will be different, so if the above is the case,
	// we'll be visiting it and not returning early.
	if strings.Contains(fileName, "ontap/api/rest/client") && strings.HasSuffix(fileName, "client.go") {
		fmt.Println("FileName: ", fileName)
		fmt.Println("Function Name: ", functionName)
		file, ok := fileMap[fileName]
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
	//fmt.Printf("Visiting function: %s in file: %s\n", functionName, fileName)

	outgoingCalls, err := findCallees(fileName, line, character)
	if err != nil {
		fmt.Println("Error finding callees: ", err)
		fmt.Println("FileName: ", fileName)
		fmt.Println("Function Name: ", functionName)
		fmt.Println("Line: ", line)
		fmt.Println("Character: ", character)
	}

	// outGoingCalls can be of length 0, indicating that we might have encountered an interface.
	if len(outgoingCalls) == 0 {
		isInterface := false

		// Get the file
		file, ok := fileMap[fileName]
		if !ok {
			panic("File not found")
			return
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

		//documentSymbols, _ := getAllTheSymbols(fileName)
		//
		//for _, documentSymbol := range documentSymbols {
		//	if documentSymbol.Kind == SymbolKindInterface {
		//		if documentSymbol.Location.Range.Start.Line < line && documentSymbol.Location.Range.End.Line > line {
		//			isInterface = true
		//			break
		//		}
		//	}
		//}

		if isInterface == true {
			implementations, err := findImplementation(fileName, line, character)
			if err != nil {
				fmt.Println("Error getting implementations: ", err)
			}
			for _, implementation := range implementations {
				if strings.Contains(implementation.Uri, "mocks") {
					continue
				}
				fileName = implementation.Uri
				fileName = strings.ReplaceAll(fileName, "file://", "")
				//wg.Add(1)
				r.traverseRecursively(fileName, implementation.Range.Start.Line, implementation.Range.Start.Character, functionName, depth+1)
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

		if strings.Contains(call.To.Detail, "github.com/netapp/trident/persistent_store") {
			continue
		}

		if strings.Contains(call.To.Detail, "github.com/netapp/trident/storage_attribute") {
			continue
		}

		if strings.Contains(call.To.Detail, "github.com/netapp/trident/storage_class") {
			continue
		}

		line = call.To.Range.Start.Line
		character = call.To.Range.Start.Character
		fileName = call.To.Uri
		fileName = strings.ReplaceAll(fileName, "file://", "")
		//wg.Add(1)
		r.traverseRecursively(fileName, line, character, call.To.Name, depth+1)
	}

	return

}

func (r *recurseTraversal) traverseRecursivelyONTAPRest(fileName string, line, character int, functionName string, depth int) {
	//defer wg.Done()
	if depth == 3 {
		return
	}

	functionID := fmt.Sprintf("%s:%d:%d:%s", fileName, line, character, functionName)

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
	// Also, functionID will be different in that case, as fileName will be different, so if the above is the case,
	// we'll be visiting it and not returning early.
	if strings.Contains(fileName, "ontap/api/rest/client") && strings.HasSuffix(fileName, "client.go") {
		fmt.Println("FileName: ", fileName)
		fmt.Println("Function Name: ", functionName)
		file, ok := fileMap[fileName]
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
	//fmt.Printf("Visiting function: %s in file: %s\n", functionName, fileName)

	outgoingCalls, err := findCallees(fileName, line, character)
	if err != nil {
		fmt.Println("Error finding callees: ", err)
		fmt.Println("FileName: ", fileName)
		fmt.Println("Function Name: ", functionName)
		fmt.Println("Line: ", line)
		fmt.Println("Character: ", character)
	}

	// outGoingCalls can be of length 0, indicating that we might have encountered an interface.
	if len(outgoingCalls) == 0 {
		documentSymbols, _ := getAllTheSymbols(fileName)
		isInterface := false
		for _, documentSymbol := range documentSymbols {
			if documentSymbol.Kind == SymbolKindInterface {
				if documentSymbol.Location.Range.Start.Line < line && documentSymbol.Location.Range.End.Line > line {
					isInterface = true
					break
				}
			}
		}

		if isInterface == true {
			implementations, err := findImplementation(fileName, line, character)
			for _, implementation := range implementations {
				if strings.Contains(implementation.Uri, "mocks") {
					continue
				}
				fileName = implementation.Uri
				fileName = strings.ReplaceAll(fileName, "file://", "")
				//wg.Add(1)
				r.traverseRecursivelyONTAPRest(fileName, implementation.Range.Start.Line, implementation.Range.Start.Character, functionName, depth+1)
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
		fileName = call.To.Uri
		fileName = strings.ReplaceAll(fileName, "file://", "")
		//wg.Add(1)
		r.traverseRecursivelyONTAPRest(fileName, line, character, call.To.Name, depth+1)
	}

	return

}

func bfsTraverseOutgoingCalls(startFileName string, startLine, startCharacter int, startFunctionName string) error {
	queue := []struct {
		fileName     string
		line         int
		character    int
		functionName string
	}{{startFileName, startLine, startCharacter, startFunctionName}}

	visited := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Construct a unique identifier for each function to avoid revisiting.
		funcID := fmt.Sprintf("%s:%s", current.fileName, current.functionName)
		if visited[funcID] {
			continue
		}
		visited[funcID] = true

		// Process the current function, e.g., print its name.
		fmt.Println("Visiting function:", current.functionName)

		outgoingCalls, err := findCallees(current.fileName, current.line, current.character)
		if err != nil {
			fmt.Println("Error finding callees for function:", current.functionName, "-", err)
			continue
		}

		for _, call := range outgoingCalls {
			if !visited[fmt.Sprintf("%s:%s", call.To.Uri, call.To.Name)] && strings.Contains(call.To.Uri, "github.com/netapp/trident") {
				// Assuming you can extract line and character from call.To.Uri or another way.
				// Add the callee to the queue.
				queue = append(queue, struct {
					fileName     string
					line         int
					character    int
					functionName string
				}{call.To.Uri, 0, 0, call.To.Name}) // Update line and character accordingly.
			}
		}
	}

	return nil
}

func readingMethodsOfOrchestratorInterface(file ast.Node) {
	var lastTypeSpec *ast.TypeSpec
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			lastTypeSpec = node // Update lastTypeSpec with the current *ast.TypeSpec node
		case *ast.InterfaceType:
			if lastTypeSpec != nil {
				if lastTypeSpec.Name.Name == "Orchestrator" {
					interfaceType := lastTypeSpec.Type.(*ast.InterfaceType)
					for _, method := range interfaceType.Methods.List {
						// method.Names represent: field/method/(type) parameter names; or nil
						// here it is just a method, so Names[0] should do
						functionRequiredMap[method.Names[0].Name] = struct{}{}
					}
				} // Print the methods of the interface
			}
		}
		return true
	})
}
