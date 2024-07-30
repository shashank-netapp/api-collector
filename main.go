package main

import (
	"context"
	"net"

	. "github.com/theshashankpal/api-collector/callgraph"
	"github.com/theshashankpal/api-collector/callgraph/lsp"
	. "github.com/theshashankpal/api-collector/logger"
)

var m = LogFields{Key: "layer", Value: "main"}

var id = 0
var workDir = "/Users/shashank/Library/CloudStorage/OneDrive-NetAppInc/Documents/Astra/TRID-POLLING-NEW/trident/"
var lspAddress = "localhost:7070"

func init() {
	Get()
}

func main() {

	ctx := context.Background()

	//Establish a TCP connection to gopls server
	Log(ctx, m).Info().Msgf("Establishing a TCP connection to gopls server at %s", lspAddress)
	conn, err := net.Dial("tcp", lspAddress)
	if err != nil {
		Log(ctx, m).Error().Msgf("Failed to establish a TCP connection to gopls server at %s", lspAddress)
		return
	}
	defer conn.Close()
	Log(ctx, m).Info().Msgf("Connection to gopls server established at %s", lspAddress)

	// Creating call-graph
	Log(ctx, m).Info().Msg("Creating a call-graph")
	var callGraph CallGraph
	callGraph = lsp.NewAbstractionLSP(ctx, conn, workDir, "trident")

	// Initialize call-graph
	Log(ctx, m).Debug().Msg("Initializing call-graph instance")
	err = callGraph.Initialize(ctx)
	if err != nil {
		Log(ctx, m).Error().Msg("Failed to initialize call-graph instance")
		return
	}
	Log(ctx, m).Info().Msg("Call-graph is created")

	Log(ctx, m).Info().Msg("Starting the traversal")

	//
	//zerolog.SetGlobalLevel(zerolog.DebugLevel)
	//log.Info().Msg("Connected to gopls server")
	//
	//// Create a new LSP instance
	//lspInstance := lsp2.NewLsp(conn, workDir)
	//
	//// Initialize the LSP
	//fmt.Println("Initializing LSP")
	//err = lspInstance.Initialize("api-collector")
	//if err != nil {
	//	fmt.Println("Failed to initialize LSP:", err)
	//	return
	//}
	//fmt.Println("LSP is Initialized")
	//traverse([]string{})
}
