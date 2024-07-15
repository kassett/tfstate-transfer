package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

var ImportIdentifierFields = []string{"id", "name"}

type ImportObject struct {
	sourceName   string
	targetName   string
	topLevelName string
	identifier   map[string]*string
}

type ImportRunResult struct {
	userDefinedResource string
	sourceResourceName  string
	targetResourceName  string
	success             bool
	errorReceived       error
	suggestion          string
}

type RunHandler struct {
	// Maps the name of the Resources the user defined
	// To all the Resources it potentially contains
	topLevelResourceMapping map[string][]string

	// Maps the names of the Resources from the source state
	// to the names of the Resources in the target state
	sourceTargetNameMapping map[string]string

	// Maps the full path of the SOURCE state resource
	// to a list of fields extracted from the state
	// that could potentially be used to import the resource
	resourceIdentifiers map[string]*ImportObject

	// The next Resources to be imported
	resourcesToImport *Stack

	// The list of imports that have completed (not necessarily successfully)
	completedImports map[string]bool

	// The results of the run for output
	importResults []ImportRunResult
}

func checkIfResourceBelongsToState(resourceName string, resourceMapping map[string]string) (bool, string, string) {
	for source, target := range resourceMapping {
		if strings.HasPrefix(resourceName, source) {
			return true, source, strings.Replace(resourceName, source, target, 1)
		}
	}
	return false, "", ""
}

func makeUnique(slice []string) *[]string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return &list
}

func NewRunHandler(stateFileContent string, resourceMapping map[string]string) *RunHandler {
	topLevelResourceMapping := make(map[string][]string)
	sourceTargetNameMapping := make(map[string]string)
	resourceIdentifiers := make(map[string]*ImportObject)
	resourcesToImport := NewStack()
	completedImports := make(map[string]bool)
	importResults := make([]ImportRunResult, 0)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stateFileContent), &parsed); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	resources, ok := parsed["resources"].([]interface{})
	if !ok {
		return nil
	}

	for _, res := range resources {
		resMap, ok := res.(map[string]interface{})
		if !ok || resMap["mode"] != "managed" {
			continue
		}

		instances, ok := resMap["instances"].([]interface{})
		if !ok {
			continue
		}

		for _, inst := range instances {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}

			attributes, ok := instMap["attributes"].(map[string]interface{})
			if !ok {
				continue
			}

			extractedFields := make(map[string]*string)
			for _, field := range ImportIdentifierFields {
				if value, ok := attributes[field].(string); ok {
					extractedFields[field] = &value
				}
			}

			fullPath := fmt.Sprintf("%s.%s", resMap["type"].(string), resMap["name"].(string))
			if index, ok := instMap["index_key"]; ok {
				if stringKey, ok := index.(string); ok {
					fullPath += fmt.Sprintf("[\"%s\"]", stringKey)
				} else {
					fullPath += fmt.Sprintf("[%v]", index)
				}
			}
			if module, ok := resMap["module"].(string); ok {
				fullPath = module + "." + fullPath
			}

			// Check if the resource belongs to something defined top-level
			belongsToState, topLevel, newFullPath := checkIfResourceBelongsToState(fullPath, resourceMapping)
			if belongsToState {
				sourceTargetNameMapping[fullPath] = newFullPath
				resourceIdentifiers[fullPath] = &ImportObject{
					sourceName:   fullPath,
					targetName:   newFullPath,
					topLevelName: topLevel,
					identifier:   extractedFields,
				}
				topLevelResourceMapping[topLevel] = append(topLevelResourceMapping[topLevel], fullPath)
			}
		}
	}

	for topLevel := range topLevelResourceMapping {
		topLevelResourceMapping[topLevel] = *makeUnique(topLevelResourceMapping[topLevel])
	}

	for initial := range sourceTargetNameMapping {
		resourcesToImport.Push(initial)
	}

	return &RunHandler{
		topLevelResourceMapping: topLevelResourceMapping,
		sourceTargetNameMapping: sourceTargetNameMapping,
		resourceIdentifiers:     resourceIdentifiers,
		resourcesToImport:       resourcesToImport,
		completedImports:        completedImports,
		importResults:           importResults,
	}
}

func (rn *RunHandler) HasNextResource() bool {
	// Check the stack to see if there are more Resources that need to be imported
	if rn.resourcesToImport.IsEmpty() {
		return false
	}
	return true
}

func (rn *RunHandler) GetTopLevelFromResource(sourceResourceName string) (string, error) {
	// Get the name of the resource called by the resource given the name of the
	// resource extracted from the state
	for parent, children := range rn.topLevelResourceMapping {
		for _, child := range children {
			if child == sourceResourceName {
				return parent, nil
			}
		}
	}
	return "", errors.New(fmt.Sprintf("resource %s was not found", sourceResourceName))
}

func (rn *RunHandler) GetNextResource() (*ImportObject, error) {
	// Get an ImportAttemptObject (essentially just a struct with all the relevant
	// info to perform an import) from the stack
	if !rn.HasNextResource() {
		return nil, errors.New("there are no more resources to import")
	}
	sourceResourceName, err := rn.resourcesToImport.Pop()
	if err != nil {
		return nil, err
	}

	return rn.resourceIdentifiers[sourceResourceName], nil

}

func (rn *RunHandler) ResourcesToDelete() []string {
	// Check which Resources can be deleted
	parentsToDelete := make([]string, 0)
	for parent, children := range rn.topLevelResourceMapping {

		allSuccessfulImports := true

		for _, child := range children {
			if !rn.completedImports[child] {
				allSuccessfulImports = false
				break
			}
		}

		if allSuccessfulImports {
			parentsToDelete = append(parentsToDelete, parent)
		}
	}

	return parentsToDelete
}

func (rn *RunHandler) ReportImportRun(sourceResourceName string,
	targetResourceName string, userDefinedResource string, errorReceived error) {
	// After having attempted to perform an import, tell the handler about the output
	importRunResult := ImportRunResult{
		userDefinedResource: userDefinedResource,
		sourceResourceName:  sourceResourceName,
		targetResourceName:  targetResourceName,
		success:             errorReceived == nil,
		errorReceived:       errorReceived,
		suggestion:          "",
	}

	rn.completedImports[sourceResourceName] = importRunResult.success
	rn.importResults = append(rn.importResults, importRunResult)
}

func (rn *RunHandler) FinishRun() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"User Resource", "Source Resource", "Target Resource", "Success", "Error"})

	// Set table style options
	table.SetBorder(true) // Set table border
	table.SetAutoWrapText(false)
	table.SetRowLine(true) // Enable row line
	table.SetColumnSeparator("│")

	redRow := []tablewriter.Colors{
		{tablewriter.FgRedColor},
		{tablewriter.FgRedColor},
		{tablewriter.FgRedColor},
		{tablewriter.FgRedColor},
		{tablewriter.FgRedColor},
	}

	greenRow := []tablewriter.Colors{
		{tablewriter.FgGreenColor},
		{tablewriter.FgGreenColor},
		{tablewriter.FgGreenColor},
		{tablewriter.FgGreenColor},
		{tablewriter.FgGreenColor},
	}

	for _, resultRow := range rn.importResults {
		success := "True"
		if !resultRow.success {
			success = "False"
		}

		errorString := "N/A"
		if resultRow.errorReceived != nil {
			errorString = resultRow.errorReceived.Error()
		}

		row := []string{
			resultRow.userDefinedResource,
			resultRow.sourceResourceName,
			resultRow.targetResourceName,
			success,
			errorString,
		}

		if resultRow.success {
			// Print row in green if success is true
			table.Rich(row, greenRow)
		} else {
			// Print row in red if success is false
			table.Rich(row, redRow)
		}
	}

	table.SetAutoWrapText(true)
	table.SetAutoFormatHeaders(true)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("")
	table.SetTablePadding("\t")

	table.Render()
}
