package internal_test

import (
	"github.com/kassett/tfstate-transfer/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshallConfigFileContent(t *testing.T) {
	configFileContent := `
{
  "sourceDir": "source",
  "targetDir": "target",
  "resources": [
    {
      "source": "aws_secretsmanager_secret.this",
      "target": "aws_secretsmanager_secret.this"
    },
    {
      "source": "aws_secretsmanager_secret_version.this",
      "target": "aws_secretsmanager_secret_version.this"
    },
    {
      "source": "aws_dynamodb_table.this",
      "target": "aws_dynamodb_table.target"
    }
  ]
}
`
	sourceDir, targetDir, resources, resourceMapping := internal.UnmarshallConfigFileContent(configFileContent)
	assert.Equal(t, "source", sourceDir, "Expected the source directory to be `source`")
	assert.Equal(t, "target", targetDir, "Expected the source directory to be `source`")
	assert.Len(t, resources, 3, "Expected 3 resources.")

	assert.Equal(t, resourceMapping["aws_dynamodb_table.this"], "aws_dynamodb_table.target")
}

func TestPullAliasesOutFromCli(t *testing.T) {
	resources := []string{
		"aws_secretsmanager_secret.this",
		"module.db:module.db2",
	}

	resources, resourceMapping := internal.PullAliasesOutFromCli(resources)
	assert.Len(t, resources, 2)
	assert.Equal(t, resourceMapping["module.db"], "module.db2")
}
