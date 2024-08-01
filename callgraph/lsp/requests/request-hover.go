package requests

import (
	"encoding/json"
	"fmt"
)

type HoverParams struct {
	TextDocumentPositionParams
	WorkDoneProgressParams
}

type HoverRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  HoverParams `json:"params"`
	ID      int         `json:"id"`

	responseChan chan map[string]interface{}
}

// need to get the structure of the reponse of the hover request
type HoverResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Result  []Location     `json:"result"`
	Error   *ResponseError `json:"error"`
}

func (r *HoverRequest) NewRequest(filePath string, line, character, id int) *HoverRequest {
	return &HoverRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/hover",
		Params: HoverParams{
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

func (r *HoverRequest) SendRequest(requestChan chan Request) {
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

func (r *HoverRequest) ReadResponse(responseChan chan *HoverResponse) {
	response := <-r.responseChan

	bytes, err := json.Marshal(response)
	if err != nil {
		// handle the error
		tempHoverResponse := &HoverResponse{
			Error: &ResponseError{
				Code:    JsonMarshalError,
				Message: fmt.Sprintf("HoverResponse #ReadResponse: failed to marshal -> %v", err),
			},
		}
		responseChan <- tempHoverResponse
		return
	}

	var hoverResponse HoverResponse
	err = json.Unmarshal(bytes, &hoverResponse)
	if err != nil {
		// handle the error
		tempHoverResponse := &HoverResponse{
			Error: &ResponseError{
				Code:    JsonUnMarshalError,
				Message: fmt.Sprintf("HoverResponse #ReadResponse: failed to unmarshal -> %v", err),
			},
		}
		responseChan <- tempHoverResponse
		return
	}

	responseChan <- &hoverResponse
}

type HoverRequestInterface interface {
	NewRequest(filePath string, line, character, id int) *HoverRequest
	SendRequest(requestChan chan Request)
	ReadResponse(responseChan chan *HoverResponse)
}
