package requests

import (
	"encoding/json"
	"fmt"
)

type CallHierarchyPrepareRequest struct {
	Jsonrpc string                     `json:"jsonrpc"`
	Method  string                     `json:"method"`
	Params  CallHierarchyPrepareParams `json:"params"`
	ID      int                        `json:"id"`

	responseChan chan map[string]interface{}
}

type CallHierarchyPrepareParams struct {
	TextDocumentPositionParams
	WorkDoneProgressParams
}

type CallHierarchyItem struct {
	/**
	 * The name of this item.
	 */
	Name string `json:"name"`

	/**
	 * The kind of this item.
	 */
	Kind SymbolKind `json:"kind"`

	/**
	 * More detail for this item, e.g. the signature of a function.
	 */
	Detail string `json:"detail"`

	/**
	 * The resource identifier of this item.
	 */
	Uri string `json:"uri"`

	/**
	 * The range enclosing this symbol not including leading/trailing whitespace
	 * but everything else, e.g. comments and code.
	 */
	Range Range `json:"range"`

	/**
	 * The range that should be selected and revealed when this symbol is being
	 * picked, e.g. the name of a function. Must be contained by the
	 * [`range`](#CallHierarchyItem.range).
	 */
	SelectionRange Range `json:"selectionRange"`
}

type CallHierarchyPrepareResponse struct {
	Jsonrpc string            `json:"jsonrpc"`
	Result  CallHierarchyItem `json:"result"`
	ID      int               `json:"id"`
	Error   *ResponseError    `json:"error"`
}

func (r *CallHierarchyPrepareRequest) NewRequest(fileName string, line, character, id int) *CallHierarchyPrepareRequest {
	textDocument := TextDocumentIdentifier{
		Uri: fmt.Sprintf("file://%s", fileName),
	}

	position := Position{Line: line, Character: character}
	textDocumentPositionParams := TextDocumentPositionParams{
		TextDocument: textDocument,
		Position:     position,
	}

	return &CallHierarchyPrepareRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/prepareCallHierarchy",
		Params: CallHierarchyPrepareParams{
			TextDocumentPositionParams: textDocumentPositionParams,
		},
		ID: id, // Example ID, ensure it is unique
	}
}

func (r *CallHierarchyPrepareRequest) SendRequest(requestChan chan Request) {
	// Form the Request
	r.responseChan = make(chan map[string]interface{})
	request := Request{
		request:      *r,
		id:           r.ID,
		responseChan: r.responseChan,
	}

	// Send the request
	requestChan <- request
}

func (r *CallHierarchyPrepareRequest) ReadResponse(callHierarchyPrepareResponseChan chan *CallHierarchyPrepareResponse) {
	response := <-r.responseChan

	bytes, err := json.Marshal(response)
	if err != nil {
		// handle error
		// Handle the error
		tempCallHierarchyPrepareResponse := &CallHierarchyPrepareResponse{
			Error: &ResponseError{
				Code:    JsonMarshalError,
				Message: fmt.Sprintf("CallHierarchyPrepareResponse #ReadResponse: failed to marshal -> %v", err),
			},
		}

		callHierarchyPrepareResponseChan <- tempCallHierarchyPrepareResponse
		return
	}

	var callHierarchyPrepareReponse CallHierarchyPrepareResponse
	err = json.Unmarshal(bytes, &callHierarchyPrepareReponse)
	if err != nil {
		// handle error
		tempCallHierarchyPrepareResponse := &CallHierarchyPrepareResponse{
			Error: &ResponseError{
				Code:    JsonUnMarshalError,
				Message: fmt.Sprintf("CallHierarchyPrepareResponse #ReadResponse: failed to unmarshal -> %v", err),
			},
		}

		callHierarchyPrepareResponseChan <- tempCallHierarchyPrepareResponse
		return
	}

	callHierarchyPrepareResponseChan <- &callHierarchyPrepareReponse
}

type CallHierarchyPrepareRequestInterface interface {
	NewRequest(fileName string, line, character, id int) *CallHierarchyPrepareRequest
	SendRequest(requestChan chan Request)
	ReadResponse(callHierarchyPrepareResponseChan chan *CallHierarchyPrepareResponse)
}
