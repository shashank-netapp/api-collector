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
	ZAPICommand  string `json:"zapi_command"`
	ONTAPCommand string `json:"ontap_command"`
	AccessLevel  string `json:"access_level"`
}

type ZAPICommandsList struct {
	Commands []ZAPICommands `json:"zapi_commands"`
}

func WriteZAPICommands(ctx context.Context, zapiCommandsMap map[string][]string, file *os.File) error {
	// Write REST APIs to a file

	zapiCommandsList := ZAPICommandsList{
		Commands: make([]ZAPICommands, 0),
	}

	for _, value := range zapiCommandsMap {
		zapiCommand := value[0]
		if _, ok := ZAPIToONTAPMapping[zapiCommand]; !ok {
			// TODO: collect all these errors and return the list of errors back.
			Log(ctx, zap).Error().Msgf("ONTAP command not found for ZAPI command %s", zapiCommand)
			continue
		}
		ontapCommand := ZAPIToONTAPMapping[zapiCommand]

		// Getting the access level
		tempParts := strings.Split(ontapCommand, " ")
		tempString := tempParts[len(tempParts)-1]
		var accessLevel string
		if tempString == "show" {
			accessLevel = "readonly"
		} else {
			accessLevel = "all"
		}

		// Creating the ZAPI command object
		tempZAPICommand := ZAPICommands{
			ZAPICommand:  value[0],
			ONTAPCommand: ontapCommand,
			AccessLevel:  accessLevel,
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

// ZAPIToONTAPMapping contains one to one mapping of ZAPI commands to ONTAP commands
// If a ZAPI command is not present in this map, please add it using the below ontap cli command
// security login role show-ontapi -ontapi <zapi command>
// this will give you the equivalent ontap command.
var ZAPIToONTAPMapping = map[string]string{
	"lun-resize":                         "lun resize",
	"volume-destroy":                     "volume destroy",
	"quota-set-entry":                    "volume quota policy rule modify",
	"snapshot-get-iter":                  "volume snapshot show",
	"iscsi-initiator-modify-chap-params": "vserver iscsi security modify",
	"volume-get-iter":                    "volume show",
	"export-policy-destroy":              "vserver export-policy delete",
	"snapmirror-release-iter":            "snapmirror release",
	"job-schedule-get-iter":              "job schedule show",
	"system-node-get-iter":               "system node show",
	"cifs-share-get-iter":                "vserver cifs share show",
	"snapmirror-get":                     "snapmirror show",
	"snapmirror-initialize":              "snapmirror initialize",
	"igroup-add":                         "lun igroup add",
	"igroup-remove":                      "lun igroup remove",
	"lun-offline":                        "lun offline",
	"volume-modify-iter-async":           "volume modify",
	"snapmirror-get-destination-iter":    "snapmirror list-destinations",
	"snapmirror-create":                  "snapmirror create",
	"net-interface-get-iter":             "network interface show",
	"aggr-space-get-iter":                "storage aggregate show-space",
	"lun-set-qos-policy-group":           "lun modify",
	"lun-get-iter":                       "lun show",
	"volume-mount":                       "volume mount",
	"snapshot-restore-volume":            "volume snapshot restore",
	"snapshot-delete":                    "volume snapshot delete",
	"qtree-modify":                       "volume qtree modify",
	"lun-create-by-size":                 "lun create",
	"volume-offline":                     "volume offline",
	"system-get-ontapi-version":          "version",
	"iscsi-initiator-delete-auth":        "vserver iscsi security delete",
	"lun-get-attribute":                  "lun show",
	"qtree-list-iter":                    "volume qtree show",
	"snapshot-create":                    "volume snapshot create",
	"snapmirror-get-iter":                "snapmirror show",
	"iscsi-initiator-get-auth":           "vserver iscsi security show",
	"lun-destroy":                        "lun delete",
	"volume-size-async":                  "volume size",
	"export-rule-destroy":                "vserver export-policy rule delete",
	"export-policy-get":                  "vserver export-policy show",
	"iscsi-initiator-add-auth":           "vserver iscsi security create",
	"iscsi-initiator-auth-get-iter":      "vserver iscsi security show",
	"snapmirror-resync":                  "snapmirror resync",
	"snapmirror-break":                   "snapmirror break",
	"igroup-get-iter":                    "lun igroup show",
	"lun-unmap":                          "lun mapping delete",
	"volume-modify-iter":                 "volume modify",
	"volume-unmount":                     "volume unmount",
	"iscsi-service-get-iter":             "vserver iscsi show",
	"snapmirror-quiesce":                 "snapmirror quiesce",
	"snapmirror-release":                 "snapmirror release",
	"clone-create":                       "volume file clone create",
	"snapmirror-policy-get-iter":         "snapmirror policy show",
	"iscsi-initiator-get-iter":           "vserver iscsi initiator show",
	"cifs-share-create":                  "vserver cifs share create",
	"lun-move":                           "lun move-in-volume",
	"quota-status":                       "volume quota show",
	"volume-rename":                      "volume rename",
	"snapmirror-update":                  "snapmirror update",
	"iscsi-initiator-get-default-auth":   "vserver iscsi show",
	"volume-create":                      "volume create",
	"quota-off":                          "volume quota off",
	"vserver-get-iter":                   "vserver show",
	"cifs-share-delete":                  "vserver cifs share delete",
	"ems-autosupport-log":                "event generate-autosupport-log",
	"lun-map-get-iter":                   "lun mapping show",
	"lun-set-attribute":                  "lun modify",
	"volume-clone-split-start":           "volume clone split start",
	"volume-clone-create":                "volume clone create",
	"qtree-rename":                       "volume qtree rename",
	"snapmirror-destroy":                 "snapmirror delete",
	"quota-list-entries-iter":            "volume quota policy rule show",
	"iscsi-node-get-name":                "vserver iscsi nodename",
	"lun-online":                         "lun online",
	"job-get-iter":                       "job show",
	"qtree-create":                       "volume qtree create",
	"export-rule-create":                 "vserver export-policy rule create",
	"snapmirror-abort":                   "snapmirror abort",
	"quota-on":                           "volume quota on",
	"export-rule-get-iter":               "vserver export-policy rule show",
	"igroup-create":                      "lun igroup create",
	"igroup-destroy":                     "lun igroup delete",
	"volume-size":                        "volume size",
	"qtree-delete-async":                 "volume qtree delete",
	"iscsi-interface-get-iter":           "vserver iscsi interface show",
	"vserver-peer-get-iter":              "vserver peer show",
	"lun-get-serial-number":              "lun serial",
	"lun-map":                            "lun mapping create",
	"volume-destroy-async":               "volume destroy",
	"volume-create-async":                "volume create",
	"volume-clone-create-async":          "volume clone create",
	"quota-resize":                       "volume quota resize",
	"iscsi-initiator-set-default-auth":   "vserver iscsi security default",
	"lun-map-list-info":                  "lun mapping show",
	"export-policy-create":               "vserver export-policy create",
	"vserver-get":                        "vserver show",
	"vserver-show-aggr-get-iter":         "vserver show-aggregates",
	"system-get-version":                 "version",
}
