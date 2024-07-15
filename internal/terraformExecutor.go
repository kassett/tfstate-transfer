package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func executeCommand(command string, directory string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Return both the error and the combined output
		return string(output), fmt.Errorf("error: %v, output: %s", err, string(output))
	}
	return string(output), nil
}
func generateStateFile(sourceDir string) string {
	command := "terraform state pull"
	output, err := executeCommand(command, sourceDir)
	if err != nil {
		fmt.Print(err)
	}
	return output
}

func runImport(targetDir string, importObject ImportObject) error {
	defaultError := errors.New("unknown error: try importing manually")

	for _, field := range ImportIdentifierFields {
		id, exists := importObject.identifier[field]
		if !exists || id == nil {
			continue
		}

		command := fmt.Sprintf("terraform import '%s' '%s'", importObject.targetName, *id)
		output, err := executeCommand(command, targetDir)

		// How to handle errors
		// If we can't import by any of our saved properties, we can't delete the state
		// If we can't import because the resource isn't importable, return nil
		// If we can't import because we've already imported, return nil

		if err != nil {
			if strings.Contains(output, "Resource already managed by Terraform") {
				return nil
			} else if strings.Contains(output, "This resource does not support import.") {
				return errors.New("resource does not implement the import protocol")
			}
			continue
		} else {
			return nil
		}
	}
	return defaultError
}

func terraformRemoveState(resource string, sourceDir string) {
	_, err := executeCommand(fmt.Sprintf("terraform state rm '%s'", resource), sourceDir)
	if err != nil {
		os.Exit(1)
	}
}
