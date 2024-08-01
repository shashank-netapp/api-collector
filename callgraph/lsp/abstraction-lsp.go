package lsp

import (
	"context"
	"net"

	. "github.com/theshashankpal/api-collector/callgraph/lsp/requests"
	. "github.com/theshashankpal/api-collector/logger"
)

var alf = LogFields{Key: "layer", Value: "abstraction-lsp"}

// AbstractionLSP Using this, I can abstract the specific LSP client behavior and implement as required.
type AbstractionLSP struct {
	lspClient LSPInterface
	name      string
}

func NewAbstractionLSP(ctx context.Context, conn net.Conn, workDir, name string) *AbstractionLSP {
	Log(ctx, alf).Trace().Msg(">>>> NewAbstractionLSP")
	defer Log(ctx, alf).Trace().Msg("<<<< NewAbstractionLSP")

	Log(ctx, alf).Debug().Msg("Instantiating new abstraction LSP")
	return &AbstractionLSP{
		lspClient: NewLsp(ctx, conn, workDir),
		name:      name,
	}
}

func (l *AbstractionLSP) Initialize(ctx context.Context) error {
	Log(ctx, alf).Trace().Msg(">>>> Initialize")
	defer Log(ctx, alf).Trace().Msg("<<<< Initialize")

	initializeRequest := InitializeRequest{}
	initializedNotification := InitializedNotification{}
	return l.lspClient.Initialize(ctx, l.name, &initializeRequest, &initializedNotification)
}

func (l *AbstractionLSP) OutgoingCalls(ctx context.Context, filePath string, line, character int) chan *CallHierarchyOutgoingCallResponse {
	Log(ctx, alf).Trace().Msg(">>>> OutgoingCalls")
	defer Log(ctx, alf).Trace().Msg("<<<< OutgoingCalls")

	callHierarchyOutgoingCallRequest := CallHierarchyOutgoingCallRequest{}
	callHierarchyPrepareRequest := CallHierarchyPrepareRequest{}
	return l.lspClient.OutgoingCalls(ctx, filePath, line, character, &callHierarchyOutgoingCallRequest, &callHierarchyPrepareRequest)
}

func (l *AbstractionLSP) Implementations(ctx context.Context, filePath string, line, character int) chan *ImplementationResponse {
	Log(ctx, alf).Trace().Msg(">>>> Implementations")
	defer Log(ctx, alf).Trace().Msg("<<<< Implementations")

	implementationRequest := ImplementationRequest{}
	return l.lspClient.Implementations(ctx, filePath, line, character, &implementationRequest)
}

func (l *AbstractionLSP) References(ctx context.Context, filePath string, line int, character int) chan *ReferenceResponse {
	Log(ctx, alf).Trace().Msg(">>>> References")
	defer Log(ctx, alf).Trace().Msg("<<<< References")

	referencesRequest := ReferenceRequest{}
	return l.lspClient.References(ctx, filePath, line, character, &referencesRequest)
}

func (l *AbstractionLSP) Hover(ctx context.Context, filePath string, line, character int) chan *HoverResponse {
	Log(ctx, alf).Trace().Msg(">>>> Hover")
	defer Log(ctx, alf).Trace().Msg("<<<< Hover")

	hoverRequest := HoverRequest{}
	return l.lspClient.Hover(ctx, filePath, line, character, &hoverRequest)
}

func (l *AbstractionLSP) DocumentSymbol(ctx context.Context, filePath string) chan *DocumentSymbolResponse {
	Log(ctx, alf).Trace().Msg(">>>> DocumentSymbol")
	defer Log(ctx, alf).Trace().Msg("<<<< DocumentSymbol")

	documentSymbolRequest := DocumentSymbolRequest{}
	return l.lspClient.DocumentSymbol(ctx, filePath, &documentSymbolRequest)
}
