package lsp

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"

	. "github.com/theshashankpal/api-collector/callgraph/lsp/requests"
	. "github.com/theshashankpal/api-collector/logger"
)

var lf = LogFields{Key: "layer", Value: "lsp"}

type LSPInterface interface {
	Initialize(ctx context.Context, name string,
		initializeRequest InitializeRequestInterface,
		initializedNotification InitializedNotificationInterface) error
	OutgoingCalls(ctx context.Context, filePath string, line, character int,
		callHierarchyOutgoingCallRequest CallHierarchyOutgoingCallRequestInterface,
		callHierarchyPrepareRequest CallHierarchyPrepareRequestInterface) chan *CallHierarchyOutgoingCallResponse
	Implementations(ctx context.Context, filePath string, line, character int,
		implementationRequest ImplementationRequestInterface) chan *ImplementationResponse
	References(ctx context.Context, filePath string, line int, character int,
		referencesRequest ReferenceRequestIntercace) chan *ReferenceResponse
	Hover(ctx context.Context, fileName string, line, character int, hoverRequest HoverRequestInterface) chan *HoverResponse
	DocumentSymbol(ctx context.Context, filePath string,
		documentSymbolRequest DocumentSymbolRequestInterface) chan *DocumentSymbolResponse
}

type LSP struct {
	conn         net.Conn
	reader       *bufio.Reader
	workdir      string
	initialized  bool
	requestChan  chan Request
	responseChan chan Request
	requester    *Requester
}

func NewLsp(ctx context.Context, conn net.Conn, workDir string) *LSP {
	Log(ctx, lf).Trace().Msg(">>>> NewLsp")
	defer Log(ctx, lf).Trace().Msg("<<<< NewLsp")

	Log(ctx, lf).Debug().Msg("Instantiating new concrete LSP")
	return &LSP{
		conn:    conn,
		workdir: workDir,
		reader:  bufio.NewReader(conn),
	}
}

func (l *LSP) Initialize(ctx context.Context, name string,
	initializeRequest InitializeRequestInterface,
	initializedNotification InitializedNotificationInterface) error {
	Log(ctx, lf).Trace().Msg(">>>> Initialize")
	defer Log(ctx, lf).Trace().Msg("<<<< Initialize")

	// Sending initialize request
	Log(ctx, lf).Debug().Str("workDir", l.workdir).Msg("Sending initialize request")
	initializeRequest = initializeRequest.NewRequest(l.workdir, name, int(uuid.New().ID()))
	err := initializeRequest.SendRequest(l.conn)
	if err != nil {
		return err
	}

	initializeResponse, err := initializeRequest.ReadResponse(l.reader)
	if err != nil {
		return err
	}

	if initializeResponse.Error != nil {
		return fmt.Errorf("#Initialize: failed to get initialize response -> %w", initializeResponse.Error)
	}

	// Sending initialized notification
	initializedNotification = initializedNotification.NewNotification()
	err = initializedNotification.SendRequest(l.conn)
	if err != nil {
		return err
	}
	err = initializedNotification.ReadResponse(l.reader)
	if err != nil {
		return err
	}

	// Start the requester after initialization.
	l.requestChan = make(chan Request, 10)
	l.responseChan = make(chan Request, 10)
	l.requester = NewRequester(l.reader, l.requestChan, l.responseChan, l.conn)

	l.initialized = true
	return nil
}

func (l *LSP) OutgoingCalls(ctx context.Context, filePath string, line, character int,
	callHierarchyOutgoingCallRequest CallHierarchyOutgoingCallRequestInterface,
	callHierarchyPrepareRequest CallHierarchyPrepareRequestInterface) chan *CallHierarchyOutgoingCallResponse {

	// Create a channel to send back the response
	callHierarchyOutgoingCallChan := make(chan *CallHierarchyOutgoingCallResponse)

	go func() {
		callHierarchyPrepareResponseChan := make(chan *CallHierarchyPrepareResponse)
		callHierarchyPrepareRequest = callHierarchyPrepareRequest.NewRequest(filePath, line, character, int(uuid.New().ID()))
		responseChan := make(chan map[string]interface{})
		go callHierarchyPrepareRequest.SendRequest(l.requestChan, responseChan)
		go callHierarchyPrepareRequest.ReadResponse(callHierarchyPrepareResponseChan, responseChan)

		// Wait for the call hierarchy prepare response
		callHierarchyPrepareResponse := <-callHierarchyPrepareResponseChan
		if callHierarchyPrepareResponse.Error != nil {
			// Send back the error
			tempCallHierarchyOutgoingCallChan := &CallHierarchyOutgoingCallResponse{
				Error: &ResponseError{
					Code:    callHierarchyPrepareResponse.Error.Code,
					Message: fmt.Sprintf("OutgoingCalls: failed to get call hierarchy prepare response -> %v", callHierarchyPrepareResponse.Error.Error()),
				},
			}

			callHierarchyOutgoingCallChan <- tempCallHierarchyOutgoingCallChan
			return
		}

		// Now need to get the actual outgoing calls
		if len(callHierarchyPrepareResponse.Result) == 0 {
			// Send back the error
			tempCallHierarchyOutgoingCallChan := &CallHierarchyOutgoingCallResponse{
				Error: &ResponseError{
					Code:    EmptyCallHierarchPrepareResponse,
					Message: fmt.Sprintf("OutgoingCalls: call hierarchy prerpare resonse is empty"),
				},
			}

			callHierarchyOutgoingCallChan <- tempCallHierarchyOutgoingCallChan
			return
		}
		callHierarchyOutgoingCallRequest = callHierarchyOutgoingCallRequest.NewRequest(callHierarchyPrepareResponse.Result[0], int(uuid.New().ID()))
		responseChan = make(chan map[string]interface{})
		go callHierarchyOutgoingCallRequest.SendRequest(l.requestChan, responseChan)
		go callHierarchyOutgoingCallRequest.ReadResponse(callHierarchyOutgoingCallChan, responseChan)
	}()

	return callHierarchyOutgoingCallChan
}

func (l *LSP) Implementations(ctx context.Context, filePath string, line, character int,
	implementationRequest ImplementationRequestInterface) chan *ImplementationResponse {

	var implementationResponseChan = make(chan *ImplementationResponse)

	implementationRequest = implementationRequest.NewRequest(filePath, line, character, int(uuid.New().ID()))

	tempResponseChan := make(chan map[string]interface{})
	go implementationRequest.SendRequest(l.requestChan, tempResponseChan)
	go implementationRequest.ReadResponse(implementationResponseChan, tempResponseChan)

	return implementationResponseChan
}

func (l *LSP) References(ctx context.Context, filePath string, line int, character int,
	referencesRequest ReferenceRequestIntercace) chan *ReferenceResponse {
	var referencesResponseChan = make(chan *ReferenceResponse)

	referencesRequest = referencesRequest.NewRequest(filePath, line, character, int(uuid.New().ID()))

	go referencesRequest.SendRequest(l.requestChan)
	go referencesRequest.ReadResponse(referencesResponseChan)

	return referencesResponseChan
}

func (l *LSP) Hover(ctx context.Context, fileName string, line, character int, hoverRequest HoverRequestInterface) chan *HoverResponse {
	var hoverResponseChan = make(chan *HoverResponse)

	hoverRequest = hoverRequest.NewRequest(fileName, line, character, int(uuid.New().ID()))

	go hoverRequest.SendRequest(l.requestChan)
	go hoverRequest.ReadResponse(hoverResponseChan)

	return hoverResponseChan
}

func (l *LSP) DocumentSymbol(ctx context.Context, filePath string,
	documentSymbolRequest DocumentSymbolRequestInterface) chan *DocumentSymbolResponse {

	var documentSymbolResponseChan = make(chan *DocumentSymbolResponse)

	documentSymbolRequest = documentSymbolRequest.NewRequest(filePath, int(uuid.New().ID()))

	go documentSymbolRequest.SendRequest(l.requestChan)
	go documentSymbolRequest.ReadResponse(documentSymbolResponseChan)

	return documentSymbolResponseChan
}
