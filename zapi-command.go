package main

import (
	"context"
	"encoding/json"
	. "github.com/theshashankpal/api-collector/logger"
	"os"
	"strings"
)

var zap = LogFields{Key: "layer", Value: "zapi"}

type ZAPICommands struct {
	FunctionName string `json:"function_name"`
	Command      string `json:"command"`
}

type ZAPICommandsList struct {
	Commands []ZAPICommands `json:"zapi_commands"`
}

func WriteZAPICommands(ctx context.Context, zapiCommandsMap map[string][]string, file *os.File) error {
	// Write REST APIs to a file

	zapiCommandsList := ZAPICommandsList{
		Commands: make([]ZAPICommands, 0),
	}

	for key, value := range zapiCommandsMap {
		functionName := strings.TrimSpace(strings.Split(key, ":")[3])
		tempZAPICommand := ZAPICommands{
			FunctionName: functionName,
			Command:      value[0],
		}
		zapiCommandsList.Commands = append(zapiCommandsList.Commands, tempZAPICommand)
	}

	jsonData, err := json.MarshalIndent(zapiCommandsList, "", "    ")
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
