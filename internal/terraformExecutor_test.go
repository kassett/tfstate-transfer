package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(value string) *string {
	return &value
}

func TestRunImportDryRunReturnsImportCommand(t *testing.T) {
	command, err := runImport("/target", ImportObject{
		targetName: "aws_dynamodb_table.this",
		identifier: map[string]*string{
			"id": strPtr("table-id"),
		},
	}, true)

	require.NoError(t, err)
	assert.Equal(t, "terraform import 'aws_dynamodb_table.this' 'table-id'", command)
}

func TestRunImportTriesNextIdentifierAfterImportFailure(t *testing.T) {
	originalRunner := runTerraformCommand
	defer func() { runTerraformCommand = originalRunner }()

	var commands []string
	runTerraformCommand = func(command string, directory string) (string, error) {
		commands = append(commands, command)
		if len(commands) == 1 {
			return "not found", errors.New("failed")
		}
		return "imported", nil
	}

	command, err := runImport("/target", ImportObject{
		targetName: "aws_kinesis_stream.this",
		identifier: map[string]*string{
			"id":   strPtr("stream-id"),
			"name": strPtr("stream-name"),
		},
	}, false)

	require.NoError(t, err)
	assert.Empty(t, command)
	assert.Equal(t, []string{
		"terraform import 'aws_kinesis_stream.this' 'stream-id'",
		"terraform import 'aws_kinesis_stream.this' 'stream-name'",
	}, commands)
}

func TestRunImportHandlesKnownTerraformErrors(t *testing.T) {
	originalRunner := runTerraformCommand
	defer func() { runTerraformCommand = originalRunner }()

	runTerraformCommand = func(command string, directory string) (string, error) {
		return "Resource already managed by Terraform", errors.New("already managed")
	}

	_, err := runImport("/target", ImportObject{
		targetName: "aws_dynamodb_table.this",
		identifier: map[string]*string{
			"id": strPtr("table-id"),
		},
	}, false)
	require.NoError(t, err)

	runTerraformCommand = func(command string, directory string) (string, error) {
		return "This resource does not support import.", errors.New("unsupported")
	}

	_, err = runImport("/target", ImportObject{
		targetName: "local_file.this",
		identifier: map[string]*string{
			"id": strPtr("file-id"),
		},
	}, false)
	require.EqualError(t, err, "resource does not implement the import protocol")
}

func TestTerraformRemoveStateDryRunReturnsCommand(t *testing.T) {
	assert.Equal(t, "terraform state rm 'aws_dynamodb_table.this'", terraformRemoveState("aws_dynamodb_table.this", "/source", true))
}
