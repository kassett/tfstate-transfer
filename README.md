# tfstate-transfer

`tfstate-transfer` is a CLI tool for transferring Terraform resources from one environment to another. 
This tool simplifies the process of moving state files and resources between different Terraform workspaces or environments, 
making it easier to manage infrastructure changes and migrations.

## Installation

To install `tfstate-transfer`, you need to have Go installed. Then, you can build and install the tool using the following commands:

```bash
go install github.com/yourusername/tfstate-transfer
```

## Usage
There are two ways to use the tfstate-transfer program. The first
of these ways is via the simple CLI interface. 
When invoking tfstate-transfer, you pass a source directory and a target directory, as follows.

```shell
tfstate-transfer -sourceDir startDirectory -targetDir endDirectory -r module.db
```
