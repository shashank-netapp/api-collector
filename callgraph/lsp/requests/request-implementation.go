package requests

import (
	"encoding/json"
	"fmt"
)

type ImplementationParams struct {
	TextDocumentPositionParams
	WorkDoneProgressParams
	PartialResultParams
}

type ImplementationRequest struct {
	Jsonrpc string               `json:"jsonrpc"`
	Method  string               `json:"method"`
	Params  ImplementationParams `json:"params"`
	ID      int                  `json:"id"`

	responseChan chan map[string]interface{}
}

type ImplementationResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Result  []Location     `json:"result"`
	Error   *ResponseError `json:"error"`
}

func (r *ImplementationRequest) NewRequest(filePath string, line, character, id int) *ImplementationRequest {
	return &ImplementationRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/implementation",
		Params: ImplementationParams{
			TextDocumentPositionParams: TextDocumentPositionParams{
				TextDocument: TextDocumentIdentifier{
					Uri: fmt.Sprintf("file://%s", filePath),
				},
				Position: Position{
					Line:      line,
					Character: character,
				},
			},
		},
		ID: id,
	}
}

func (r *ImplementationRequest) SendRequest(requestChan chan Request) {
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

func (r *ImplementationRequest) ReadResponse(responseChan chan *ImplementationResponse) {
	response := <-r.responseChan

	bytes, err := json.Marshal(response)
	if err != nil {
		// handle error
		tempImplementationResponse := &ImplementationResponse{
			Error: &ResponseError{
				Code:    JsonMarshalError,
				Message: fmt.Sprintf("ImplementationResponse #ReadResponse: failed to marshal -> %v", err),
			},
		}
		responseChan <- tempImplementationResponse
		return
	}

	var implementationResponse ImplementationResponse
	err = json.Unmarshal(bytes, &implementationResponse)
	if err != nil {
		// handle error
		tempImplementationResponse := &ImplementationResponse{
			Error: &ResponseError{
				Code:    JsonUnMarshalError,
				Message: fmt.Sprintf("ImplementationResponse #ReadResponse: failed to unmarshal -> %v", err),
			},
		}
		responseChan <- tempImplementationResponse
		return
	}

	responseChan <- &implementationResponse
}

type ImplementationRequestInterface interface {
	NewRequest(filePath string, line, character, id int) *ImplementationRequest
	SendRequest(requestChan chan Request)
	ReadResponse(responseChan chan *ImplementationResponse)
}
