package ast_traverser

import (
	"context"
	"sync"

	. "github.com/theshashankpal/api-collector/callgraph"
	. "github.com/theshashankpal/api-collector/logger"
	. "github.com/theshashankpal/api-collector/traverser/ast-traverser/dfs"
)

var tf = LogFields{Key: "layer", Value: "traverser"}

// Search in an interface which will be implemented by DfsTraverser/BfsTraverser.
type Search interface {
	Initialize(ctx context.Context, done chan bool)
	Traverse(ctx context.Context, mapChan chan map[string][]string)
}

type AstTraverser struct {
	workDir       string
	callGraph     CallGraph
	zapi          bool
	rest          bool
	restTraverser Search
	zapiTraverser Search
}

func NewAstTraverser(workDir string, callGraph CallGraph, rest bool, zapi bool) *AstTraverser {
	return &AstTraverser{
		workDir:   workDir,
		callGraph: callGraph,
		rest:      rest,
		zapi:      zapi,
	}
}

func (t *AstTraverser) Initialize(ctx context.Context) {

	// Because we are using the same callGraph for both REST and ZAPI recursers, we need to lock it
	callGraphMU := new(sync.Mutex)
	if t.rest {
		Log(ctx, tf).Debug().Msg("Creating a new REST recurser")
		t.restTraverser = NewDfsTraverser(t.callGraph, callGraphMU, t.workDir, RESTRecurserType)
	}

	if t.zapi {
		Log(ctx, tf).Debug().Msg("Creating a new ZAPI recurser")
		t.zapiTraverser = NewDfsTraverser(t.callGraph, callGraphMU, t.workDir, ZAPIRecurserType)
	}

	var (
		restTraverserInitialized chan bool
		zapiTraverserInitialized chan bool
	)

	if t.rest {
		restTraverserInitialized = make(chan bool)
		go t.restTraverser.Initialize(ctx, restTraverserInitialized)
	}

	if t.zapi {
		zapiTraverserInitialized = make(chan bool)
		go t.zapiTraverser.Initialize(ctx, zapiTraverserInitialized)
	}

	if t.rest {
		<-restTraverserInitialized
	}

	if t.zapi {
		<-zapiTraverserInitialized
	}

	Log(ctx, tf).Debug().Msg("Initialization of traversers was successful")
}

func (t *AstTraverser) Traverse(ctx context.Context) (chan map[string][]string, chan map[string][]string) {
	restAPIsMapChan := make(chan map[string][]string)
	zapiCommandsMapChan := make(chan map[string][]string)

	if t.rest {
		t.restTraverser.Traverse(ctx, restAPIsMapChan)
	}

	if t.zapi {
		t.zapiTraverser.Traverse(ctx, zapiCommandsMapChan)
	}

	return restAPIsMapChan, zapiCommandsMapChan
}
