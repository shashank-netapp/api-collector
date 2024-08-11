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
	API         string `json:"api"`
	Method      string `json:"method"`
	AccessLevel string `json:"access_level"`
}

type RestAPIsList struct {
	APIs []RestAPIs `json:"apis"`
}

func WriteRESTAPIs(ctx context.Context, restAPIsMap map[string][]string, file *os.File) error {
	// Write REST APIs to a file

	aggregatedMethodsMap := aggregateMethods(restAPIsMap)

	restAPIsList := RestAPIsList{
		APIs: make([]RestAPIs, 0),
	}

	for key, value := range aggregatedMethodsMap {
		api := key

		// Getting the access level
		var accessLevel string
		aggregatedMethods := value

		switch {
		case contains(aggregatedMethods, "DELETE"):
			accessLevel = "all"
		case contains(aggregatedMethods, "PATCH") && contains(aggregatedMethods, "POST"):
			accessLevel = "read_create_modify"
		case contains(aggregatedMethods, "PATCH"):
			accessLevel = "read_modify"
		case contains(aggregatedMethods, "POST"):
			accessLevel = "read_create"
		default:
			accessLevel = "readonly"
		}

		tempRestAPIs := RestAPIs{
			Method:      strings.Join(aggregatedMethods, ", "),
			API:         api,
			AccessLevel: accessLevel,
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

func contains(list []string, word string) bool {
	for _, item := range list {
		if item == word {
			return true
		}
	}
	return false
}

func aggregateMethods(restAPIsMap map[string][]string) map[string][]string {
	apiSet := make(map[string][]string)
	for _, value := range restAPIsMap {
		api := value[1]
		method := value[0]
		if _, ok := apiSet[api]; !ok {
			apiSet[api] = []string{method}
		} else {
			apiSet[api] = append(apiSet[api], method)
		}
	}

	return apiSet
}
