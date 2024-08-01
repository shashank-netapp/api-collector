package callgraph

import (
	"context"
	"github.com/theshashankpal/api-collector/callgraph/lsp/requests"
)

type CallGraph interface {
	Initialize(ctx context.Context) error
	OutgoingCalls(ctx context.Context, filePath string, line, character int) chan *requests.CallHierarchyOutgoingCallResponse
	Implementations(ctx context.Context, filePath string, line, character int) chan *requests.ImplementationResponse
	References(ctx context.Context, filePath string, line int, character int) chan *requests.ReferenceResponse
	Hover(ctx context.Context, filePath string, line, character int) chan *requests.HoverResponse
	DocumentSymbol(ctx context.Context, filePath string) chan *requests.DocumentSymbolResponse
}
