package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunExecutesTerraformCommandsWhenNotDryRun(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	originalRunner := runTerraformCommand
	defer func() { runTerraformCommand = originalRunner }()

	var commands []string
	runTerraformCommand = func(command string, directory string) (string, error) {
		commands = append(commands, command)
		if command == "terraform state pull" {
			assert.Equal(t, sourceDir, directory)
			return stateWithManagedResources, nil
		}
		return "", nil
	}

	Run(sourceDir, targetDir, map[string]string{
		"aws_dynamodb_table.this": "aws_dynamodb_table.this",
	}, false)

	assert.Equal(t, []string{
		"terraform state pull",
		"terraform import 'aws_dynamodb_table.this' 'table-id'",
		"terraform state rm 'aws_dynamodb_table.this'",
	}, commands)
}

func TestRunUsesDryRunCommandsWithoutExecutingImportOrRemove(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	originalRunner := runTerraformCommand
	defer func() { runTerraformCommand = originalRunner }()

	var commands []string
	runTerraformCommand = func(command string, directory string) (string, error) {
		commands = append(commands, command)
		if command == "terraform state pull" {
			return stateWithManagedResources, nil
		}
		t.Fatalf("unexpected command execution in dry-run mode: %s", command)
		return "", nil
	}

	Run(sourceDir, targetDir, map[string]string{
		"aws_dynamodb_table.this": "aws_dynamodb_table.this",
	}, true)

	assert.Equal(t, []string{"terraform state pull"}, commands)
}

func TestOpenConfigFileReadsContent(t *testing.T) {
	configFile, err := os.CreateTemp(t.TempDir(), "config-*.json")
	require.NoError(t, err)
	_, err = configFile.WriteString(`{"sourceDir":"source","targetDir":"target","resources":[]}`)
	require.NoError(t, err)
	require.NoError(t, configFile.Close())

	assert.JSONEq(t, `{"sourceDir":"source","targetDir":"target","resources":[]}`, OpenConfigFile(configFile.Name()))
}
