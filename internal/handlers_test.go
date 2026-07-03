package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stateWithManagedResources = `{
  "resources": [
    {
      "mode": "managed",
      "type": "aws_dynamodb_table",
      "name": "this",
      "instances": [
        {
          "attributes": {
            "id": "table-id",
            "name": "table-name"
          }
        }
      ]
    },
    {
      "mode": "managed",
      "module": "module.table_count[0]",
      "type": "aws_dynamodb_table",
      "name": "this",
      "instances": [
        {
          "attributes": {
            "id": "module-table-id"
          }
        }
      ]
    },
    {
      "mode": "managed",
      "type": "aws_secretsmanager_secret",
      "name": "iterate_foreach",
      "instances": [
        {
          "index_key": "one",
          "attributes": {
            "id": "secret-one"
          }
        }
      ]
    },
    {
      "mode": "data",
      "type": "aws_caller_identity",
      "name": "current",
      "instances": [
        {
          "attributes": {
            "id": "ignored"
          }
        }
      ]
    }
  ]
}`

func TestNewRunHandlerExtractsManagedResourcesAndAliases(t *testing.T) {
	handler := NewRunHandler(stateWithManagedResources, map[string]string{
		"aws_dynamodb_table.this":                   "aws_dynamodb_table.renamed",
		"module.table_count[0]":                     "module.table_count[2]",
		"aws_secretsmanager_secret.iterate_foreach": "aws_secretsmanager_secret.iterate_foreach",
	})

	require.NotNil(t, handler)
	assert.Equal(t, "aws_dynamodb_table.renamed", handler.sourceTargetNameMapping["aws_dynamodb_table.this"])
	assert.Equal(t, "module.table_count[2].aws_dynamodb_table.this", handler.sourceTargetNameMapping["module.table_count[0].aws_dynamodb_table.this"])
	assert.Equal(t, "aws_secretsmanager_secret.iterate_foreach[\"one\"]", handler.sourceTargetNameMapping["aws_secretsmanager_secret.iterate_foreach[\"one\"]"])
	assert.NotContains(t, handler.sourceTargetNameMapping, "data.aws_caller_identity.current")

	table := handler.resourceIdentifiers["aws_dynamodb_table.this"]
	require.NotNil(t, table)
	require.NotNil(t, table.identifier["id"])
	require.NotNil(t, table.identifier["name"])
	assert.Equal(t, "table-id", *table.identifier["id"])
	assert.Equal(t, "table-name", *table.identifier["name"])
}

func TestNewRunHandlerReturnsNilForInvalidState(t *testing.T) {
	assert.Nil(t, NewRunHandler(`not-json`, map[string]string{}))
	assert.Nil(t, NewRunHandler(`{"version":4}`, map[string]string{}))
}

func TestResourcesToDeleteOnlyIncludesFullyCompletedParents(t *testing.T) {
	handler := NewRunHandler(stateWithManagedResources, map[string]string{
		"aws_dynamodb_table.this":                   "aws_dynamodb_table.this",
		"aws_secretsmanager_secret.iterate_foreach": "aws_secretsmanager_secret.iterate_foreach",
	})
	require.NotNil(t, handler)

	handler.completedImports["aws_dynamodb_table.this"] = true
	handler.completedImports["aws_secretsmanager_secret.iterate_foreach[\"one\"]"] = false

	assert.ElementsMatch(t, []string{"aws_dynamodb_table.this"}, handler.ResourcesToDelete())
}

func TestReportImportRunRecordsSuccessAndFailure(t *testing.T) {
	handler := NewRunHandler(stateWithManagedResources, map[string]string{
		"aws_dynamodb_table.this": "aws_dynamodb_table.this",
	})
	require.NotNil(t, handler)

	handler.ReportImportRun("aws_dynamodb_table.this", "aws_dynamodb_table.this", "aws_dynamodb_table.this", nil)
	handler.ReportImportRun("missing", "missing", "missing", errors.New("failed"))

	assert.True(t, handler.completedImports["aws_dynamodb_table.this"])
	assert.False(t, handler.completedImports["missing"])
	require.Len(t, handler.importResults, 2)
	assert.True(t, handler.importResults[0].success)
	assert.False(t, handler.importResults[1].success)
}
