package requests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/theshashankpal/api-collector/utils"
)

type Request struct {
	request      interface{}
	id           int
	responseChan chan map[string]interface{}
}

type Requester struct {
	reader           *bufio.Reader
	neededRequests   map[int]struct{}
	neededRequestsMU *sync.Mutex
	cachedResponses  map[int]map[string]interface{}
}

func NewRequester(reader *bufio.Reader, requestChan chan Request, responseChan chan Request, conn net.Conn) *Requester {
	requester := &Requester{
		reader:           reader,
		neededRequests:   make(map[int]struct{}),
		cachedResponses:  make(map[int]map[string]interface{}),
		neededRequestsMU: new(sync.Mutex),
	}

	// Start the go routines
	go requester.submitRequest(requestChan, responseChan, conn)
	go requester.readResponse(responseChan)

	return requester
}

func (r *Requester) submitRequest(request chan Request, responseReader chan Request, conn net.Conn) {
	for {
		select {
		case req := <-request:
			// Send the request
			if req.request == nil {
				// bad request handle it
			} else {
				// Send the request
				requestJSON, err := json.Marshal(req.request)
				requestWithHeader := utils.ConstructRequest(requestJSON)
				_, err = fmt.Fprintf(conn, requestWithHeader)
				if err != nil {
					fmt.Printf("error sending request with request id %d : %v\n", req.id, err)
				}

				// Save it in the needed map
				r.neededRequestsMU.Lock()
				r.neededRequests[req.id] = struct{}{}
				r.neededRequestsMU.Unlock()

				// Now signal readReponse go routine that you can try to read the response of this request.
				responseReader <- req
			}
		}
	}
}

func (r *Requester) readResponse(responseReader chan Request) {
	for {
		select {
		case req := <-responseReader:
			// Read the response
			for {
				// Check whether we have the response for this request cached or not?
				if _, ok := r.cachedResponses[req.id]; ok {
					req.responseChan <- r.cachedResponses[req.id]
					// Now we've served the response, remove it from the cache and needed map
					delete(r.cachedResponses, req.id)
					r.neededRequestsMU.Lock()
					delete(r.neededRequests, req.id)
					r.neededRequestsMU.Unlock()
					break
				}

				contentLength, err := utils.FindTheContentLength(r.reader)
				if err != nil {
					fmt.Printf("Error finding content length of request id %d : %v\n", req.id, err)
				}

				content := make([]byte, contentLength)
				_, err = io.ReadFull(r.reader, content)
				if err != nil {
					fmt.Printf("Error reading response of request id %d : %v\n", req.id, err)
				}

				// Unmarshal JSON content
				var message map[string]interface{}
				err = json.Unmarshal(content, &message)
				if err != nil {
					fmt.Printf("Error unmarshalling response of request id %d : %v\n", req.id, err)
				}

				if _, ok := message["id"]; !ok {
					// not of need, ignore it
					continue
				}

				id := int(message["id"].(float64))

				// Check whether we even require this response or not
				r.neededRequestsMU.Lock()
				if _, ok := r.neededRequests[id]; !ok {
					// not of need, ignore it
					r.neededRequestsMU.Unlock()
					continue
				}
				r.neededRequestsMU.Unlock()

				// See whether this response is of the request we are looking for
				if id == req.id {
					req.responseChan <- message
					r.neededRequestsMU.Lock()
					delete(r.neededRequests, req.id)
					r.neededRequestsMU.Unlock()
					break
				} else {
					// cache this response
					r.cachedResponses[id] = message
				}
			}
		}
	}
}
