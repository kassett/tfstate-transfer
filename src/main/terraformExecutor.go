package main

import (
	"fmt"
	"os/exec"
)

func executeCommand(command string, directory string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
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

func runImport(targetDir string, newResourceName string, stateResource *StateResource) error {
	for _, field := range ImportIdentifierFields {
		identifier, exists := stateResource.resourceFields[field]
		if !exists || identifier == nil {
			continue
		}
		command := fmt.Sprintf("terraform import \"%s\" \"%s\"", newResourceName, *identifier)
		_, err := executeCommand(command, targetDir)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to import resource")
}
