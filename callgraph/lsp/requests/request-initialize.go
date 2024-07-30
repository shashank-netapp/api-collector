package requests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/theshashankpal/api-collector/utlis"
	"io"
	"net"
)

// InitializeError provides additional information about initialization errors
type InitializeError struct {
	Retry bool `json:"retry"`
}

// ResponseError represents an error in the initialization process
type InitializeResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    InitializeError `json:"data,omitempty"`
}

func (i *InitializeResponseError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, retry: %t", i.Code, i.Message, i.Data.Retry)
}

// InitializeResult represents the successful initialization response
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// InitializeResponse represents the structure of the response from the server
type InitializeResponse struct {
	Jsonrpc string                   `json:"jsonrpc"`
	ID      int                      `json:"id"`
	Result  InitializeResult         `json:"result,omitempty"`
	Error   *InitializeResponseError `json:"error,omitempty"`
}

type InitializeRequest struct {
	Jsonrpc string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  InitializeParams `json:"params"`
	ID      int              `json:"id"`

	responseChan chan map[string]interface{}
}

type InitializeParams struct {
	ProcessID             *int               `json:"processId"`
	RootPath              *string            `json:"rootPath,omitempty"`
	RootURI               string             `json:"rootUri"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	Trace                 string             `json:"trace,omitempty"`
	WorkspaceFolders      []WorkspaceFolder  `json:"workspaceFolders,omitempty"`
}

func (r *InitializeRequest) NewRequest(workDir, name string, id int) *InitializeRequest {
	return &InitializeRequest{
		Jsonrpc: "2.0",
		Method:  "initialize",
		Params: InitializeParams{
			ProcessID: nil,
			WorkspaceFolders: []WorkspaceFolder{
				{
					URI:  fmt.Sprintf("file://%s", workDir),
					Name: name,
				},
			},
		},
		ID: id,
	}
}

func (r *InitializeRequest) SendRequest(conn net.Conn) error {
	requestJSON, err := json.Marshal(r)
	if err != nil {
		return err
	}

	request := utlis.ConstructRequest(requestJSON)

	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		return err
	}

	return nil
}

func (r *InitializeRequest) ReadResponse(reader *bufio.Reader) (*InitializeResponse, error) {
	for {
		contentLength, err := utlis.FindTheContentLength(reader)
		if err != nil {
			return nil, err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(reader, content)
		if err != nil {
			return nil, err
		}

		var initializeResponse InitializeResponse
		err = json.Unmarshal(content, &initializeResponse)
		if err != nil {
			return nil, err
		}

		if initializeResponse.ID == r.ID {
			return &initializeResponse, nil
		}
	}
}

type InitializeRequestInterface interface {
	NewRequest(workDir, name string, id int) *InitializeRequest
	SendRequest(conn net.Conn) error
	ReadResponse(reader *bufio.Reader) (*InitializeResponse, error)
}
