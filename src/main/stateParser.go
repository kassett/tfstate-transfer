package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

var ImportIdentifierFields = []string{"id", "name"}

type StateResource struct {
	fullPath       string
	resourceFields map[string]*string
}

// StateHolder holds resources in a nested map structure.
type StateHolder struct {
	resources map[string]StateResource
}

func (sh *StateHolder) AddResource(fullPath string, importableFields map[string]*string) {
	sh.resources[fullPath] = StateResource{
		fullPath:       fullPath,
		resourceFields: importableFields,
	}
}

func (sh *StateHolder) GetResource(resourcePath string) StateResource {
	return sh.resources[resourcePath]
}

func (sh *StateHolder) GetResourcesByParent(resourcePrefix string) []StateResource {
	subResources := make([]StateResource, 0)
	for fullPath, resource := range sh.resources {
		if strings.HasPrefix(fullPath, resourcePrefix) {
			subResources = append(subResources, resource)
		}
	}
	return subResources
}

func NewStateHolder(stateFileContent string) *StateHolder {
	stateHolder := &StateHolder{
		resources: make(map[string]StateResource),
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stateFileContent), &parsed); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return stateHolder
	}

	resources, ok := parsed["resources"].([]interface{})
	if !ok {
		return stateHolder
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
				fullPath += fmt.Sprintf("[%v]", index)
			}
			if module, ok := resMap["module"].(string); ok {
				fullPath = module + "." + fullPath
			}

			stateHolder.AddResource(fullPath, extractedFields)
		}
	}

	return stateHolder
}
