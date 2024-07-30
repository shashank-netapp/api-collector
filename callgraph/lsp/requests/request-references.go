package requests

import (
	"encoding/json"
	"fmt"
)

type ReferenceRequest struct {
	Jsonrpc string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  ReferencesParams `json:"params"`
	ID      int              `json:"id"`

	responseChan chan map[string]interface{}
}

type ReferencesParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

type ReferenceResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	Result  []Location     `json:"result"`
	ID      int            `json:"id"`
	Error   *ResponseError `json:"error"`
}

func (r *ReferenceRequest) NewRequest(filePath string, line int, character int, id int) *ReferenceRequest {
	textDocument := TextDocumentIdentifier{
		Uri: fmt.Sprintf("file://%s", filePath),
	}

	position := Position{Line: line, Character: character}
	textDocumentPositionParams := TextDocumentPositionParams{
		TextDocument: textDocument,
		Position:     position,
	}

	return &ReferenceRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/references",
		Params: ReferencesParams{
			TextDocumentPositionParams: textDocumentPositionParams,
			Context:                    ReferenceContext{IncludeDeclaration: false},
		},
		ID: id, // Example ID, ensure it is unique
	}
}

func (r *ReferenceRequest) SendRequest(requestChan chan Request) {
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

func (r *ReferenceRequest) ReadResponse(responseChan chan *ReferenceResponse) {
	response := <-r.responseChan

	bytes, err := json.Marshal(response)
	if err != nil {
		// handle the error
		tempReferenceResponse := &ReferenceResponse{
			Error: &ResponseError{
				Code:    JsonMarshalError,
				Message: fmt.Sprintf("ReferenceResponse #ReadResponse: failed to marshal -> %v", err),
			},
		}
		responseChan <- tempReferenceResponse
		return
	}

	var referenceResponse ReferenceResponse
	err = json.Unmarshal(bytes, &referenceResponse)
	if err != nil {
		// handle the error
		tempReferenceResponse := &ReferenceResponse{
			Error: &ResponseError{
				Code:    JsonUnMarshalError,
				Message: fmt.Sprintf("ReferenceResponse #ReadResponse: failed to unmarshal -> %v", err),
			},
		}
		responseChan <- tempReferenceResponse
		return
	}

	responseChan <- &referenceResponse
}

type ReferenceRequestIntercace interface {
	NewRequest(filePath string, line int, character int, id int) *ReferenceRequest
	SendRequest(requestChan chan Request)
	ReadResponse(responseChan chan *ReferenceResponse)
}
