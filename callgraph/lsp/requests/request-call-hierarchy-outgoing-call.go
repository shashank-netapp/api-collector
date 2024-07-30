package requests

import (
	"encoding/json"
	"fmt"
)

type CallHierarchyOutgoingCallRequest struct {
	Jsonrpc string                           `json:"jsonrpc"`
	Method  string                           `json:"method"`
	Params  CallHierarchyOutgoingCallsParams `json:"params"`
	ID      int                              `json:"id"`

	responseChan chan map[string]interface{}
}

type CallHierarchyOutgoingCallsParams struct {
	WorkDoneProgressParams
	PartialResultParams
	Item CallHierarchyItem `json:"item"`
}

type CallHierarchyOutgoingCall struct {

	/**
	 * The item that is called.
	 */
	To CallHierarchyItem `json:"to"`

	/**
	 * The range at which this item is called. This is the range relative to
	 * the caller, e.g the item passed to `callHierarchy/outgoingCalls` request.
	 */
	FromRanges []Range `json:"fromRanges"`
}

type CallHierarchyOutgoingCallResponse struct {
	Jsonrpc string                      `json:"jsonrpc"`
	Result  []CallHierarchyOutgoingCall `json:"result"`
	ID      int                         `json:"id"`
	Error   *ResponseError              `json:"error"`
}

func (r *CallHierarchyOutgoingCallRequest) NewRequest(callHierarchyItem CallHierarchyItem, id int) *CallHierarchyOutgoingCallRequest {
	return &CallHierarchyOutgoingCallRequest{
		Jsonrpc: "2.0",
		Method:  "callHierarchy/outgoingCalls",
		Params: CallHierarchyOutgoingCallsParams{
			Item: callHierarchyItem,
		},
		ID: id, // Example ID, ensure it is unique
	}
}

func (r *CallHierarchyOutgoingCallRequest) SendRequest(requestChan chan Request) {
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

func (r *CallHierarchyOutgoingCallRequest) ReadResponse(callHierarchyOutgoingCallResponseChan chan *CallHierarchyOutgoingCallResponse) {
	response := <-r.responseChan

	bytes, err := json.Marshal(response)
	if err != nil {
		// Handle the error
		tempCallHierarchyOutgoingCallResponse := &CallHierarchyOutgoingCallResponse{
			Error: &ResponseError{
				Code:    JsonMarshalError,
				Message: fmt.Sprintf("CallHierarchyOutgoingCallRequest #ReadResponse: failed to marshal -> %v", err),
			},
		}

		callHierarchyOutgoingCallResponseChan <- tempCallHierarchyOutgoingCallResponse
		return
	}

	var callHierarchyOutgoingCallResponse CallHierarchyOutgoingCallResponse
	err = json.Unmarshal(bytes, &callHierarchyOutgoingCallResponse)
	if err != nil {
		// Handle the error
		tempCallHierarchyOutgoingCallResponse := &CallHierarchyOutgoingCallResponse{
			Error: &ResponseError{
				Code:    JsonUnMarshalError,
				Message: fmt.Sprintf("CallHierarchyOutgoingCallRequest #ReadResponse: failed to unmarshal -> %v", err),
			},
		}

		callHierarchyOutgoingCallResponseChan <- tempCallHierarchyOutgoingCallResponse
		return
	}

	callHierarchyOutgoingCallResponseChan <- &callHierarchyOutgoingCallResponse
}

type CallHierarchyOutgoingCallRequestInterface interface {
	NewRequest(callHierarchyItem CallHierarchyItem, id int) *CallHierarchyOutgoingCallRequest
	SendRequest(requestChan chan Request)
	ReadResponse(callHierarchyOutgoingCallResponseChan chan *CallHierarchyOutgoingCallResponse)
}
