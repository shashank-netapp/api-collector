package bfs

import (
	"github.com/theshashankpal/api-collector/loader"
	"sync"
)

type BfsTraverser struct {
	pack []*loader.Package
	wg   *sync.WaitGroup
}

//func bfsTraverseOutgoingCalls(startFileName string, startLine, startCharacter int, startFunctionName string) error {
//	queue := []struct {
//		fileName     string
//		line         int
//		character    int
//		functionName string
//	}{{startFileName, startLine, startCharacter, startFunctionName}}
//
//	visited := make(map[string]bool)
//
//	for len(queue) > 0 {
//		current := queue[0]
//		queue = queue[1:]
//
//		// Construct a unique identifier for each function to avoid revisiting.
//		funcID := fmt.Sprintf("%s:%s", current.fileName, current.functionName)
//		if visited[funcID] {
//			continue
//		}
//		visited[funcID] = true
//
//		// Process the current function, e.g., print its name.
//		fmt.Println("Visiting function:", current.functionName)
//
//		outgoingCalls, err := findCallees(current.fileName, current.line, current.character)
//		if err != nil {
//			fmt.Println("Error finding callees for function:", current.functionName, "-", err)
//			continue
//		}
//
//		for _, call := range outgoingCalls {
//			if !visited[fmt.Sprintf("%s:%s", call.To.Uri, call.To.Name)] && strings.Contains(call.To.Uri, "github.com/netapp/trident") {
//				// Assuming you can extract line and character from call.To.Uri or another way.
//				// Add the callee to the queue.
//				queue = append(queue, struct {
//					fileName     string
//					line         int
//					character    int
//					functionName string
//				}{call.To.Uri, 0, 0, call.To.Name}) // Update line and character accordingly.
//			}
//		}
//	}
//
//	return nil
//}
