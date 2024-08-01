package dfs

import (
	"context"
	. "github.com/theshashankpal/api-collector/callgraph"
	"github.com/theshashankpal/api-collector/loader"
	. "github.com/theshashankpal/api-collector/logger"
	. "github.com/theshashankpal/api-collector/traverser/ast-traverser/dfs/recurser"
	"go/ast"
	"go/token"
	"strings"
	"sync"
)

type RecurserType int

const (
	RESTRecurserType RecurserType = iota
	ZAPIRecurserType
)

func (r RecurserType) String() string {
	switch r {
	case RESTRecurserType:
		return "RESTRecurser"
	case ZAPIRecurserType:
		return "ZAPIRecurser"
	default:
		return "Unknown"
	}
}

var dfsF = LogFields{Key: "layer", Value: "dfs-traverser"}

type recurser interface {
	Traverse(ctx context.Context, restAPIsMapChan chan map[string][]string)
	SetFileSet(fset *token.FileSet)
	SetFileMap(fileMap map[string]*ast.File)
	SetPackages(pkgs []*loader.Package)
}

type DfsTraverser struct {
	recurser
	workDir      string
	initialized  bool
	recurserType RecurserType
}

func NewDfsTraverser(callGraph CallGraph, callGraphMu *sync.Mutex, workDir string, recurserType RecurserType) *DfsTraverser {
	switch recurserType {
	case RESTRecurserType:
		return &DfsTraverser{
			workDir:      workDir,
			recurser:     NewRESTRecurser(callGraph, callGraphMu),
			recurserType: recurserType,
		}
	case ZAPIRecurserType:
		return &DfsTraverser{
			workDir:      workDir,
			recurser:     NewZAPIRecurser(callGraph, callGraphMu),
			recurserType: recurserType,
		}
	default:
		return nil
	}
}

func (d *DfsTraverser) Initialize(ctx context.Context, done chan bool) {
	// Load the package
	Log(ctx, dfsF).Debug().
		Stringer("recurserType", d.recurserType).
		Msgf("Loading packages at work direcrory : %s", d.workDir)

	pack, err := loader.LoadRoots(d.workDir)
	if err != nil {
		Log(ctx, dfsF).Panic().Stack().
			Stringer("recurserType", d.recurserType).
			Msgf("Error : %s", err)
	}
	d.SetPackages(pack)

	Log(ctx, dfsF).Debug().
		Stringer("recurserType", d.recurserType).
		Msgf("Loading AST syntax for packages at `github.com/netapp/trident/storage_drivers/ontap/api`")
	fileMap := make(map[string]*ast.File)
	for _, pkg := range pack {
		// Adding only the package which contains api calls
		if strings.Contains(pkg.PkgPath, "github.com/netapp/trident/storage_drivers/ontap/api") {
			pkg.NeedSyntax()
			for _, file := range pkg.Syntax {
				fileMap[pkg.Fset.File(file.Package).Name()] = file
			}
		}
	}
	d.SetFileMap(fileMap)
	Log(ctx, dfsF).Debug().
		Stringer("recurserType", d.recurserType).
		Msgf("AST syntax at `github.com/netapp/trident/storage_drivers/ontap/api` loaded successfully")

	d.initialized = true
	done <- true
}

func (d *DfsTraverser) Traverse(ctx context.Context, mapChan chan map[string][]string) {
	if d.initialized {
		d.recurser.Traverse(context.Background(), mapChan)
	} else {
		Log(context.Background(), dfsF).Panic().Stack().Msg("DfsTraverser not initialized")
	}
}
