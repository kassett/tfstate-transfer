package main

import (
	"fmt"
	"strings"
)

func getParentChildMapping(resourceMapping map[string]string, stateHolder *StateHolder) map[string][]string {
	parentChildMapping := make(map[string][]string)
	for originalResource, _ := range resourceMapping {
		for _, childResource := range stateHolder.GetResourcesByParent(originalResource) {
			parentChildMapping[originalResource] = append(parentChildMapping[originalResource], childResource.fullPath)
		}
	}
	return parentChildMapping
}

func transferState(
	resourceMapping map[string]string,
	stateHolder *StateHolder,
	sourceDir string,
	targetDir string,
	dryRun bool,
) {
	parentChildMapping := getParentChildMapping(resourceMapping, stateHolder)
	for parent, children := range parentChildMapping {

		for _, child := range children {
			resource := stateHolder.GetResource(child)
			newResourceName := strings.Replace(child, parent, resourceMapping[parent], 1)
			err := runImport(targetDir, newResourceName, &resource)
			if err != nil {
				fmt.Print(err)
			}
		}

		fmt.Print(parent)
	}

}

func main() {
	sourceDir, targetDir, resourceMapping, dryRun := parseArguments()
	stateFileContent := generateStateFile(sourceDir)
	stateHolder := NewStateHolder(stateFileContent)

	transferState(resourceMapping, stateHolder, sourceDir, targetDir, dryRun)
}
