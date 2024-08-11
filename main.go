package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	. "github.com/theshashankpal/api-collector/callgraph"
	. "github.com/theshashankpal/api-collector/callgraph/lsp"
	. "github.com/theshashankpal/api-collector/logger"
	. "github.com/theshashankpal/api-collector/traverser"
	. "github.com/theshashankpal/api-collector/traverser/ast-traverser"
)

var m = LogFields{Key: "layer", Value: "main"}

var (
	rest              = flag.Bool("rest", false, "Scrape REST api endpoints")
	zapi              = flag.Bool("zapi", false, "Scrape ZAPI commands")
	workDir           = flag.String("work_dir", "", "Absolute path of the root of the Trident")
	goplsAddress      = flag.String("gopls", "", "Address where the GOPLS server is running")
	logLevel          = flag.String("log_level", "info", "Provide the level for logger, default is INFO")
	restAPIOutputFile = flag.String("rest_out", "rest_apis.json", "Output file for REST APIs, json format")
	zapiOutputFile    = flag.String("zapi_out", "zapi_commands.json", "Output file for ZAPI commands, json format")
)

func main() {

	ctx := context.Background()

	flag.Parse()

	Logger(*logLevel)

	if err := validateFlags(); err != nil {
		Log(ctx, m).Error().Msg(err.Error())
		return
	}

	flag.Visit(printFlag)

	workDirTraverser := getWorkDirTraverser(*workDir)

	//Establish a TCP connection to gopls server
	Log(ctx, m).Info().Msgf("Establishing a TCP connection to gopls server at %s", *goplsAddress)
	conn, err := net.Dial("tcp", *goplsAddress)
	if err != nil {
		Log(ctx, m).Error().Err(err).Msgf("Failed to establish a TCP connection to gopls server at %s", *goplsAddress)
		return
	}
	defer conn.Close()
	Log(ctx, m).Info().Msgf("Connection to gopls server established at %s", *goplsAddress)

	// Creating call-graph
	Log(ctx, m).Info().Msg("Creating a call-graph")
	var callGraph CallGraph
	callGraph = NewAbstractionLSP(ctx, conn, *workDir, "trident")

	// Initialize call-graph
	Log(ctx, m).Debug().Msg("Initializing call-graph instance")
	err = callGraph.Initialize(ctx)
	if err != nil {
		Log(ctx, m).Error().Err(err).Msg("Failed to initialize call-graph instance")
		return
	}
	Log(ctx, m).Info().Msg("Call-graph is created")

	// Creating a traverser and initializing it.
	Log(ctx, m).Info().Msg("Creating a new traverser")
	var traverser Traverser
	traverser = NewAstTraverser(workDirTraverser, callGraph, *rest, *zapi)
	Log(ctx, m).Info().Msg("Traverser created")

	Log(ctx, m).Info().Msg("Initializing traverser")
	traverser.Initialize(ctx)
	Log(ctx, m).Info().Msg("Traverser initialized")

	// Traversing
	Log(ctx, m).Info().
		Str("workDir", *workDir).
		Bool("zapi", *zapi).
		Bool("rest", *rest).
		Msg("Traversing...")
	restAPIsMapChan, zapiCommandsMapChan := traverser.Traverse(ctx)

	//Log(ctx, m).Info().Msg("Traversing completed, writing to files")
	tempWg := new(sync.WaitGroup)
	if *rest {
		tempWg.Add(1)
		go func() {
			defer tempWg.Done()
			file, err := os.Create(*restAPIOutputFile)
			if err != nil {
				Log(ctx, m).Error().Err(err).Msgf("Failed to create file %s", *restAPIOutputFile)
				return
			}

			restAPIsMap := <-restAPIsMapChan
			Log(ctx, rst).Info().Msgf("Retrieved restAPIsMap from the channel")
			Log(ctx, m).Info().Msgf("Writing REST APIs to the file :%s", *restAPIOutputFile)
			err = WriteRESTAPIs(ctx, restAPIsMap, file)
			if err != nil {
				Log(ctx, m).Error().Err(err).Msgf("Failed to write REST APIs to file %s", *restAPIOutputFile)
				return
			}
			Log(ctx, m).Info().Msgf("REST APIs written to the file")
		}()
	}

	if *zapi {
		tempWg.Add(1)
		go func() {
			defer tempWg.Done()
			file, err := os.Create(*zapiOutputFile)
			if err != nil {
				Log(ctx, m).Error().Err(err).Msgf("Failed to create file %s", *zapiOutputFile)
				return
			}

			zapiCommandsMap := <-zapiCommandsMapChan
			Log(ctx, rst).Info().Msgf("Retrieved zapiCommandMap from the channel")
			Log(ctx, m).Info().Msgf("Writing ZAPI commands to the file :%s", *zapiOutputFile)
			err = WriteZAPICommands(ctx, zapiCommandsMap, file)
			if err != nil {
				Log(ctx, m).Error().Err(err).Msgf("Failed to write ZAPI commands to file %s", *zapiOutputFile)
				return
			}
			Log(ctx, m).Info().Msgf("ZAPI Commands are written to the file")
		}()
	}

	tempWg.Wait()
}

func printFlag(f *flag.Flag) {
	Log(context.Background(), m).Debug().
		Str("Flag", f.Name).
		Str("Value", f.Value.String()).
		Send()
}

func getWorkDirTraverser(path string) string {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	path = path + "/..."
	return path
}

func validateFlags() error {
	if *rest == false && *zapi == false {
		return fmt.Errorf("at least one of the flags -rest or -zapi must be set")
	}

	if *workDir == "" {
		return fmt.Errorf("flag -work_dir must be set")
	}

	if *goplsAddress == "" {
		return fmt.Errorf("flag -gopls must be set")
	}

	return nil
}
