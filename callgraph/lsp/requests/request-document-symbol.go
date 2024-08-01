package requests

import (
	"encoding/json"
	"fmt"
)

type DocumentSymbolParams struct {
	WorkDoneProgressParams
	PartialResultParams
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbolRequest struct {
	Jsonrpc string               `json:"jsonrpc"`
	Method  string               `json:"method"`
	Params  DocumentSymbolParams `json:"params"`
	ID      int                  `json:"id"`

	responseChan chan map[string]interface{}
}

type DocumentSymbolResult struct {
	/**
	 * The name of this symbol.
	 */
	Name string `json:"name"`

	/**
	 * The kind of this symbol.
	 */
	Kind SymbolKind `json:"kind"`

	/**
	 * Indicates if this symbol is deprecated.
	 *
	 * @deprecated Use tags instead
	 */
	Deprecated bool `json:"deprecated,omitempty"`

	/**
	 * The location of this symbol. The location's range is used by a tool
	 * to reveal the location in the editor. If the symbol is selected in the
	 * tool the range's start information is used to position the cursor. So
	 * the range usually spans more then the actual symbol's name and does
	 * normally include things like visibility modifiers.
	 *
	 * The range doesn't have to denote a node range in the sense of an abstract
	 * syntax tree. It can therefore not be used to re-construct a hierarchy of
	 * the symbols.
	 */
	Location Location `json:"location"`

	/**
	 * The name of the symbol containing this symbol. This information is for
	 * user interface purposes (e.g. to render a qualifier in the user interface
	 * if necessary). It can't be used to re-infer a hierarchy for the document
	 * symbols.
	 */
	ContainerName string `json:"containerName"`
}

type DocumentSymbolResponse struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Result  []DocumentSymbolResult `json:"result"`
	ID      int                    `json:"id"`
	Error   *ResponseError         `json:"error"`
}

func (r *DocumentSymbolRequest) NewRequest(filePath string, id int) *DocumentSymbolRequest {
	return &DocumentSymbolRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/documentSymbol",
		Params: DocumentSymbolParams{
			TextDocument: TextDocumentIdentifier{
				Uri: fmt.Sprintf("file://%s", filePath),
			},
		},
		ID: id,
	}
}

func (r *DocumentSymbolRequest) SendRequest(requestChan chan Request) {
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

func (r *DocumentSymbolRequest) ReadResponse(responseChan chan *DocumentSymbolResponse) {
	response := <-r.responseChan

	bytes, err := json.Marshal(response)
	if err != nil {
		// handle error
		tempDocumentSymbolResponse := &DocumentSymbolResponse{
			Error: &ResponseError{
				Code:    JsonMarshalError,
				Message: fmt.Sprintf("DocumentSymbolResponse #ReadResponse: failed to marshal -> %v", err),
			},
		}
		responseChan <- tempDocumentSymbolResponse
		return
	}

	var documentSymbolResponse DocumentSymbolResponse
	err = json.Unmarshal(bytes, &documentSymbolResponse)
	if err != nil {
		// handle error
		tempDocumentSymbolResponse := &DocumentSymbolResponse{
			Error: &ResponseError{
				Code:    JsonUnMarshalError,
				Message: fmt.Sprintf("DocumentSymbolResponse #ReadResponse: failed to unmarshal -> %v", err),
			},
		}
		responseChan <- tempDocumentSymbolResponse
		return
	}

	responseChan <- &documentSymbolResponse
}

type DocumentSymbolRequestInterface interface {
	NewRequest(filePath string, id int) *DocumentSymbolRequest
	SendRequest(requestChan chan Request)
	ReadResponse(responseChan chan *DocumentSymbolResponse)
}
