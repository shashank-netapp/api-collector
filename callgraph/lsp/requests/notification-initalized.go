package requests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/theshashankpal/api-collector/utlis"
)

type InitializedParams struct {
}

type InitializedNotification struct {
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  InitializedParams `json:"params"`
	ID      int               `json:"id"`
}

func (r *InitializedNotification) NewNotification() *InitializedNotification {
	return &InitializedNotification{
		Jsonrpc: "2.0",
		Method:  "initialized",
		Params:  InitializedParams{},
	}
}

func (r *InitializedNotification) SendRequest(conn net.Conn) error {
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

func (r *InitializedNotification) ReadResponse(reader *bufio.Reader) error {
	for {
		contentLength, err := utlis.FindTheContentLength(reader)
		if err != nil {
			return err
		}

		content := make([]byte, contentLength)
		_, err = io.ReadFull(reader, content)
		if err != nil {
			return err
		}

		fmt.Println(string(content))

		// Check if the response contains a specific message indicating completion
		// This condition will vary based on your server's response format
		if strings.Contains(string(content), `"method": "server/ready"`) || strings.Contains(string(content), `"packagesLoaded": true`) || strings.Contains(string(content), "Finished loading packages.") {
			// Server has indicated that it's ready or packages are loaded
			break
		}
	}

	return nil
}

type InitializedNotificationInterface interface {
	NewNotification() *InitializedNotification
	SendRequest(conn net.Conn) error
	ReadResponse(reader *bufio.Reader) error
}
