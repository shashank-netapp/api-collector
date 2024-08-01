package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	. "github.com/theshashankpal/api-collector/logger"
)

var uf = LogFields{Key: "layer", Value: "utils"}

func FindTheContentLength(reader *bufio.Reader) (int, error) {
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

func ConstructRequest(requestJSON []byte) string {
	contentLength := len(requestJSON)
	header := fmt.Sprintf("Content-Length: %d\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n", contentLength)
	return header + string(requestJSON)
}

func FindTehDifference(ctx context.Context) {
	file1, err := os.OpenFile("api.txt", os.O_RDONLY, 0644)
	if err != nil {
		Log(ctx, uf).Fatal().Msgf("failed opening file: %s", err)
	}
	defer file1.Close()

	file2, err := os.OpenFile("api_Copy.txt", os.O_RDONLY, 0644)
	if err != nil {
		Log(ctx, uf).Fatal().Msgf("failed opening file: %s", err)
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

	missingElements := FindMissingElements(list2, list1)
	fmt.Println("Missing elements from list2 in list1:", missingElements)

	missingElements = FindMissingElements(list1, list2)
	fmt.Println("Missing elements from list1 in list2:", missingElements)

}

func FindMissingElements(list1, list2 []string) []string {
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

//func AppendTofile(ctx context.Context, text string, f os.File) error {
//	// Open file in append mode, create it if it does not exist, open in write-only mode
//	//file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	//if err != nil {
//	//	log.Fatalf("failed opening file: %s", err)
//	//}
//	//defer file.Close()
//
//	// Write text to file
//	_, err := f.WriteString(text + "\n") // Adding a newline for each text added
//	if err != nil {
//		Log(ctx, uf).Error().Msgf("failed writing to file: %s", err)
//		return err
//	}
//}

func TruncateFile(ctx context.Context, filePath string) {
	// Open the file in write-only mode with the O_TRUNC flag to truncate it to zero length
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		Log(ctx, uf).Error().Msgf("failed truncating file: %s", err)
	}
	defer file.Close()
}
