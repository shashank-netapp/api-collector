package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func getImplementation(fileName string, line, character, id int, conn net.Conn) ([]Location, error) {
	requestImplementation := newImplementationRequest(fileName, line, character, id)

	requestImplementationJSON, err := json.Marshal(requestImplementation)
	if err != nil {
		return nil, err
	}

	request := createTheRequest(requestImplementationJSON)
	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		return nil, err
	}
	readers := bufio.NewReader(conn)

	for {

		contentLength, err := findTheContentLength(readers)
		if err != nil {
			return nil, err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(readers, content)
		if err != nil {
			return nil, err
		}

		var JSONResponse = struct {
			Jsonrpc string     `json:"jsonrpc"`
			Result  []Location `json:"result"`
			ID      int        `json:"id"`
		}{}

		// Unmarshal JSON content
		err = json.Unmarshal(content, &JSONResponse)
		if err != nil {
			return nil, err
		}

		if JSONResponse.ID == id {
			return JSONResponse.Result, nil
		}
	}
}

func getOutgoingCalls(callHierarchyItem *CallHierarchyItem, id int, conn net.Conn) ([]CallHierarchyOutgoingCall, error) {

	requestOutgoingCalls := CallHierarchyOutgoingCallRequest{
		Jsonrpc: "2.0",
		Method:  "callHierarchy/outgoingCalls",
		Params: CallHierarchyOutgoingCallsParams{
			Item: *callHierarchyItem,
		},
		ID: id, // Example ID, ensure it is unique
	}

	requestOutgoingCallsJSON, err := json.Marshal(requestOutgoingCalls)
	if err != nil {
		return nil, err
	}

	request := createTheRequest(requestOutgoingCallsJSON)
	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		return nil, err

	}
	readers := bufio.NewReader(conn)

	for {
		contentLength, err := findTheContentLength(readers)
		if err != nil {
			return nil, err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(readers, content)
		if err != nil {
			return nil, err
		}

		var JSONRepsnse = struct {
			Jsonrpc string                      `json:"jsonrpc"`
			Result  []CallHierarchyOutgoingCall `json:"result"`
			ID      int                         `json:"id"`
		}{}

		// Unmarshal JSON content
		err = json.Unmarshal(content, &JSONRepsnse)
		if err != nil {
			return nil, err
		}

		if JSONRepsnse.ID == id {
			return JSONRepsnse.Result, nil
		}

	}
}

func prepareCallHierarchy(fileName string, line int, character int, id int, conn net.Conn) (*CallHierarchyItem, error) {
	textDocument := TextDocumentIdentifier{
		Uri: fmt.Sprintf("file://%s", fileName),
	}

	position := Position{Line: line, Character: character}
	textDocumentPositionParams := TextDocumentPositionParams{
		TextDocument: textDocument,
		Position:     position,
	}

	requestCallHierarchy := CallHierarchyPrepareRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/prepareCallHierarchy",
		Params: CallHierarchyPrepareParams{
			TextDocumentPositionParams: textDocumentPositionParams,
		},
		ID: id, // Example ID, ensure it is unique
	}

	requestCallHierarchyJSON, err := json.Marshal(requestCallHierarchy)
	if err != nil {
		fmt.Println("Failed to marshal request:", err)
		return nil, err
	}

	request := createTheRequest(requestCallHierarchyJSON)
	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return nil, err
	}
	readers := bufio.NewReader(conn)

	for {
		contentLength, err := findTheContentLength(readers)
		if err != nil {
			fmt.Print(err)
			return nil, err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(readers, content)
		if err != nil {
			fmt.Println("Failed to read content:", err)
			return nil, err
		}

		var JSONRepsnse = struct {
			Jsonrpc string              `json:"jsonrpc"`
			Result  []CallHierarchyItem `json:"result"`
			ID      int                 `json:"id"`
		}{}

		// Unmarshal JSON content
		err = json.Unmarshal(content, &JSONRepsnse)
		if err != nil {
			fmt.Println("Failed to unmarshal JSON:", err)
			return nil, err
		}

		if JSONRepsnse.ID == id {
			if len(JSONRepsnse.Result) == 0 {
				return nil, fmt.Errorf("no call hierarchy items found")
			} else {
				return &JSONRepsnse.Result[0], nil
			}
		}

	}

}

func findTheContentLength(reader *bufio.Reader) (int, error) {
	var headers string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break // End of headers
			}
			fmt.Println("Failed to read:", err)
			return -1, err
		}
		if line == "\r\n" { // End of headers
			break
		}
		headers += line
	}

	var contentLength int
	var err error
	for _, header := range strings.Split(headers, "\r\n") {
		if strings.HasPrefix(header, "Content-Length:") {
			parts := strings.Split(header, ":")
			if len(parts) > 1 {
				contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
				if err != nil {
					fmt.Println("Invalid Content-Length:", err)
					return -1, err
				}
				break
			}
		}
	}

	if contentLength <= 0 {
		return -1, fmt.Errorf("Content-Length not found or invalid")
	}

	return contentLength, nil
}

// Create a function to send the 'initialized' notification
func sendInitializedNotification(conn net.Conn) error {
	request := struct {
		Jsonrpc string            `json:"jsonrpc"`
		Method  string            `json:"method"`
		Params  InitializedParams `json:"params"`
	}{
		Jsonrpc: "2.0",
		Method:  "initialized",
		Params:  InitializedParams{},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}

	contentLength := len(requestJSON)
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", contentLength)

	fullRequest := header + string(requestJSON)

	_, err = fmt.Fprintf(conn, fullRequest)
	if err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	reader := bufio.NewReader(conn)

	for {
		contentLength, err := findTheContentLength(reader)
		if err != nil {
			return err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(reader, content)
		if err != nil {
			return fmt.Errorf("error reading content: %w", err)
		}

		fmt.Println("Received message:", string(content))

		// Check if the response contains a specific message indicating completion
		// This condition will vary based on your server's response format
		if strings.Contains(string(content), `"method": "server/ready"`) || strings.Contains(string(content), `"packagesLoaded": true`) || strings.Contains(string(content), "Finished loading packages.") {
			// Server has indicated that it's ready or packages are loaded
			break
		}
	}

	return nil
}

func sendInitializeRequest(conn net.Conn, workDir string) error {
	// Construct the initialize request
	initParams := InitializeParams{
		ProcessID: nil,
		WorkspaceFolders: []WorkspaceFolder{
			{
				URI:  fmt.Sprintf("file://%s", workDir),
				Name: "trident",
			},
		},
	}
	request := InitializeRequest{
		Jsonrpc: "2.0",
		Method:  "initialize",
		Params:  initParams,
		ID:      1, // Example ID
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	contentLength := len(requestJSON)

	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", contentLength)

	fullRequest := header + string(requestJSON)

	// Send the request
	_, err = fmt.Fprintf(conn, fullRequest)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(conn)

	contentLength, err = findTheContentLength(reader)
	if err != nil {
		return err
	}

	response, err := readTheResponse(contentLength, reader)

	fmt.Println(response)

	return nil
}

func readTheResponse(contentLength int, reader *bufio.Reader) (string, error) {
	content := make([]byte, contentLength)
	_, err := io.ReadFull(reader, content)
	if err != nil {
		fmt.Println("Failed to read content:", err)
		return "", nil
	}

	// Unmarshal JSON content
	var message map[string]interface{}
	err = json.Unmarshal(content, &message)
	if err != nil {
		fmt.Println("Failed to unmarshal JSON:", err)
		return "", nil
	}

	prettyJSON, err := json.MarshalIndent(message, "", "    ") // Indent with four spaces
	if err != nil {
		fmt.Println("Failed to marshal JSON:", err)
		return "", nil
	}

	// Convert the byte slice to a string and print
	return string(prettyJSON), err
}

func references(fileName string, line int, character int, id int, conn net.Conn) ([]Location, error) {
	textDocument := TextDocumentIdentifier{
		Uri: fmt.Sprintf("file://%s", fileName),
	}

	position := Position{Line: line, Character: character}
	textDocumentPositionParams := TextDocumentPositionParams{
		TextDocument: textDocument,
		Position:     position,
	}

	requestReference := ReferenceRequest{
		Jsonrpc: "2.0",
		Method:  "textDocument/references",
		Params: ReferencesParams{
			TextDocumentPositionParams: textDocumentPositionParams,
			Context:                    ReferenceContext{IncludeDeclaration: false},
		},
		ID: id, // Example ID, ensure it is unique
	}

	requestReferenceJSON, err := json.Marshal(requestReference)
	if err != nil {
		return nil, err
	}

	request := createTheRequest(requestReferenceJSON)
	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		return nil, err
	}
	readers := bufio.NewReader(conn)

	contentLength, err := findTheContentLength(readers)
	if err != nil {
		return nil, err
	}

	content := make([]byte, contentLength)
	_, err = io.ReadFull(readers, content)
	if err != nil {
		return nil, err
	}

	var JSONResponse = struct {
		Jsonrpc string     `json:"jsonrpc"`
		Result  []Location `json:"result"`
		ID      int        `json:"id"`
	}{}

	// Unmarshal JSON content
	err = json.Unmarshal(content, &JSONResponse)
	if err != nil {
		return nil, err
	}

	return JSONResponse.Result, nil
}

func createTheRequest(requestJson []byte) string {
	contentLength := len(requestJson)

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", contentLength)

	fullRequest := header + string(requestJson)

	return fullRequest

}

func findAllTheSymbols(fileName string, id int, conn net.Conn) ([]DocumentSymbolReponse, error) {
	requestFindAllSymbol := newDocumentSymbolRequest(fileName, id)

	requestFindAllSymbolJSON, err := json.Marshal(requestFindAllSymbol)
	if err != nil {
		return nil, err
	}

	request := createTheRequest(requestFindAllSymbolJSON)
	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		return nil, err
	}
	readers := bufio.NewReader(conn)

	for {

		contentLength, err := findTheContentLength(readers)
		if err != nil {
			return nil, err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(readers, content)
		if err != nil {
			return nil, err
		}

		var JSONResponse = struct {
			Jsonrpc string                  `json:"jsonrpc"`
			Result  []DocumentSymbolReponse `json:"result"`
			ID      int                     `json:"id"`
		}{}

		// Unmarshal JSON content
		err = json.Unmarshal(content, &JSONResponse)
		if err != nil {
			return nil, err
		}

		if JSONResponse.ID == id {
			return JSONResponse.Result, nil
		}
	}

}

func hover(fileName string, line, character, id int) (string, error) {
	requestFindAllSymbol := newHoverRequest(fileName, line, character, id)

	requestFindAllSymbolJSON, err := json.Marshal(requestFindAllSymbol)
	if err != nil {
		return "", err
	}

	request := createTheRequest(requestFindAllSymbolJSON)
	// Send the request
	_, err = fmt.Fprintf(conn, request)
	if err != nil {
		return "", err
	}
	readers := bufio.NewReader(conn)

	contentLength, err := findTheContentLength(readers)
	if err != nil {
		return "", err
	}

	response, err := readTheResponse(contentLength, readers)
	if err != nil {
		return "", err
	}

	return response, nil

}
