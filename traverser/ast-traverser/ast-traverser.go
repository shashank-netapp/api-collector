package ast_traverser

import (
	"context"
	"go/ast"
	"strings"
	"sync"

	. "github.com/theshashankpal/api-collector/callgraph"
	"github.com/theshashankpal/api-collector/loader"
	. "github.com/theshashankpal/api-collector/logger"
)

var tf = LogFields{Key: "layer", Value: "traverser"}

var dir = "/Users/shashank/Library/CloudStorage/OneDrive-NetAppInc/Documents/Astra/TRID-POLLING-NEW/trident/storage_drivers/ontap/api/..."
var visited = make(map[string]struct{})
var wg sync.WaitGroup
var mutex sync.Mutex
var fileMap = make(map[string]*ast.File)
var apiMap = make(map[string][]string)

type Search interface {
	traverseZAPI(ctx context.Context)
	traverseREST(ctx context.Context)
}

type AstTraverser struct {
	workDir   string
	fileMap   map[string]*ast.File
	callGraph CallGraph
	search    Search
	pack      []*loader.Package
}

func NewAstTraverser(workDir string, callGraph CallGraph, search Search) *AstTraverser {
	return &AstTraverser{
		workDir:   workDir,
		fileMap:   make(map[string]*ast.File),
		callGraph: callGraph,
		search:    search,
	}
}

func (t *AstTraverser) LoadPackages(ctx context.Context) error {
	// Load the package.
	Log(ctx, tf).Debug().Msgf("Loading packages at work direcrory : %s", t.workDir)
	pack, err := loader.LoadRoots(t.workDir)
	if err != nil {
		return err
	}

	Log(ctx, tf).Debug().Msgf("Loading ast syntax for packages at `github.com/netapp/trident/storage_drivers/ontap/api`")
	for _, pkg := range pack {
		// Adding only the package which contains api calls
		if strings.Contains(pkg.PkgPath, "github.com/netapp/trident/storage_drivers/ontap/api") {
			pkg.NeedSyntax()
			for _, file := range pkg.Syntax {
				t.fileMap[pkg.Fset.File(file.Package).Name()] = file
			}
		}
	}

	t.pack = pack

	Log(ctx, tf).Debug().Msgf("Packages at `github.com/netapp/trident/storage_drivers/ontap/api` loaded successfully")
	return nil
}

func (t *AstTraverser) Traverse(ctx context.Context, traverseREST, traverseZAPI bool) {

	if traverseREST {
		go t.search.traverseREST(ctx)
	}

	if traverseZAPI {
		go t.search.traverseZAPI(ctx)
	}

	// Wait here using wait group

}
