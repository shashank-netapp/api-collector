package main

import (
	"context"
	"encoding/json"
	. "github.com/theshashankpal/api-collector/logger"
	"os"
	"strings"
)

var rst = LogFields{Key: "layer", Value: "rest"}

type RestAPIs struct {
	FunctionName string `json:"function_name"`
	API          string `json:"api"`
	Method       string `json:"method"`
}

type RestAPIsList struct {
	APIs []RestAPIs `json:"apis"`
}

func WriteRESTAPIs(ctx context.Context, restAPIsMap map[string][]string, file *os.File) error {
	// Write REST APIs to a file

	restAPIsList := RestAPIsList{
		APIs: make([]RestAPIs, 0),
	}

	for key, value := range restAPIsMap {
		functionName := strings.TrimSpace(strings.Split(key, ":")[3])
		tempRestAPIs := RestAPIs{
			FunctionName: functionName,
			Method:       value[0],
			API:          value[1],
		}
		restAPIsList.APIs = append(restAPIsList.APIs, tempRestAPIs)
	}

	jsonData, err := json.MarshalIndent(restAPIsList, "", "    ")
	if err != nil {
		return err
	}

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
