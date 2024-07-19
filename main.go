package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var conn net.Conn
var err error
var id = 0

func main() {
	// Establish a TCP connection to gopls server

	//findTehDifference()

	conn, err = net.Dial("tcp", "localhost:7070")
	if err != nil {
		fmt.Println("Failed to connect to gopls server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Sending initialize request")
	err = sendInitializeRequest(conn, "/Users/shashank/Library/CloudStorage/OneDrive-NetAppInc/Documents/Astra/TRID-POLLING-NEW/trident/")

	fmt.Println("Sending initialized notification")

	err = sendInitializedNotification(conn)
	if err != nil {
		fmt.Println("Failed to send initialized notification:", err)
		return
	}
	fmt.Println("Initialized")
	traverse([]string{})
	//
	//http.HandleFunc("/callees", findCallees)
	//http.HandleFunc("/reference", findReferences)
	//http.HandleFunc("/implementation", findImplementation)
	//
	//fmt.Println("Starting server on localhost:9000")
	//err = http.ListenAndServe(":9000", nil)
	//if err != nil {
	//	fmt.Println("Error starting server: ", err)
	//}

}

func findCallees(fileName string, line, character int) ([]CallHierarchyOutgoingCall, error) {
	// Parse the query parameters for function name and file name
	//fileName := r.URL.Query().Get("file")
	//line := r.URL.Query().Get("line")
	//character := r.URL.Query().Get("character")
	//
	//// Check if function name and file name are provided
	//if fileName == "" || line == "" || character == "" {
	//	http.Error(w, "Missing file name or line or character", http.StatusBadRequest)
	//	return
	//}
	//
	//// Convert the line and character to integers
	//lineInt, err := strconv.Atoi(line)
	//if err != nil {
	//	http.Error(w, "Invalid line number", http.StatusBadRequest)
	//	return
	//}
	//
	//characterInt, err := strconv.Atoi(character)
	//if err != nil {
	//	http.Error(w, "Invalid character number", http.StatusBadRequest)
	//	return
	//}

	id++
	callHierarchyItem, err := prepareCallHierarchy(fileName, line, character, id, conn)
	if err != nil {
		//fmt.Println("Failed to prepare call hierarchy:", err)
		//http.Error(w, "Failed to prepare call hierarchy", http.StatusInternalServerError)
		return nil, err
	}

	id++
	outgoinCalls, err := getOutgoingCalls(callHierarchyItem, id, conn)
	if err != nil {
		//fmt.Println("Failed to get outgoing calls:", err)
		//http.Error(w, "Failed to get outgoing calls", http.StatusInternalServerError)
		return nil, err
	}

	// Convert the list of callers to a JSON array
	//outgoinCallsJSON, err := json.Marshal(findCalleesResponse{Callees: outgoinCalls})
	//if err != nil {
	//	//http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
	//	return nil,err
	//}

	return outgoinCalls, nil

	// Write the JSON array to the response
	//w.Header().Set("Content-Type", "application/json")
	//w.Write(outgoinCallsJSON)
}

type findCalleesResponse struct {
	Callees []CallHierarchyOutgoingCall `json:"callees"`
}

func findReferences(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters for function name and file name
	fileName := r.URL.Query().Get("file")
	line := r.URL.Query().Get("line")
	character := r.URL.Query().Get("character")

	// Check if function name and file name are provided
	if fileName == "" || line == "" || character == "" {
		http.Error(w, "Missing file name or line or character", http.StatusBadRequest)
		return
	}

	// Convert the line and character to integers
	lineInt, err := strconv.Atoi(line)
	if err != nil {
		http.Error(w, "Invalid line number", http.StatusBadRequest)
		return
	}

	characterInt, err := strconv.Atoi(character)
	if err != nil {
		http.Error(w, "Invalid character number", http.StatusBadRequest)
		return
	}

	id++
	references, err := references(fileName, lineInt, characterInt, id, conn)
	if err != nil {
		fmt.Println("Failed to prepare call hierarchy:", err)
		http.Error(w, "Failed to prepare call hierarchy", http.StatusInternalServerError)
		return
	}

	// Convert the list of callers to a JSON array
	referencesJSON, err := json.Marshal(referencesResponse{References: references})
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	// Write the JSON array to the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(referencesJSON)
}

type referencesResponse struct {
	References []Location `json:"references"`
}

func findImplementation(fileName string, line, character int) ([]Location, error) {

	id++
	implementation, err := getImplementation(fileName, line, character, id, conn)
	if err != nil {
		return nil, err
	}

	return implementation, nil
}

type implementationResponse struct {
	Implementations []Location `json:"locations"`
}

func getAllTheSymbols(fileName string) ([]DocumentSymbolReponse, error) {
	id++
	symbols, err := findAllTheSymbols(fileName, id, conn)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

func findTehDifference() {
	file1, err := os.OpenFile("api.txt", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file1.Close()

	file2, err := os.OpenFile("api_Copy.txt", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file2.Close()

	var list1 []string
	var list2 []string
	map1 := make(map[string][]string)
	map2 := make(map[string][]string)

	scanner := bufio.NewScanner(file1)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		temp := []string{strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])}
		key := strings.TrimSpace(parts[0])
		map1[key] = temp
	}

	scanner = bufio.NewScanner(file2)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		temp := []string{strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])}
		key := strings.TrimSpace(parts[0])
		map2[key] = temp
	}

	for key, _ := range map1 {
		list1 = append(list1, key)
	}

	for key, _ := range map2 {
		list2 = append(list2, key)
	}

	missingElements := findMissingElements(list2, list1)
	fmt.Println("Missing elements from list2 in list1:", missingElements)

	missingElements = findMissingElements(list1, list2)
	fmt.Println("Missing elements from list1 in list2:", missingElements)

}

func findMissingElements(list1, list2 []string) []string {
	elementMap := make(map[string]bool)
	missingElements := []string{}

	// Mark elements from list1 in the map
	for _, item := range list1 {
		elementMap[item] = true
	}

	// Check for missing elements from list2 in list1
	for _, item := range list2 {
		if _, found := elementMap[item]; !found {
			missingElements = append(missingElements, item)
		}
	}

	return missingElements
}
