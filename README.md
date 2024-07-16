# tfstate-transfer

`tfstate-transfer` is a CLI tool for transferring Terraform resources from one environment to another. 
<<<<<<< HEAD
This tool simplifies the process of moving state files and resources between different Terraform workspaces or environments, 
making it easier to manage infrastructure changes and migrations.

## Installation

To install `tfstate-transfer`, you need to have Go installed. Then, you can build and install the tool using the following commands:

```bash
go install github.com/yourusername/tfstate-transfer
```

## Usage
There are two ways to use the tfstate-transfer program. The first
=======
This tool eases the pain that developers experience when moving to a new Terraform environment, requiring
tediously editing a tfstate file or manually writing hundreds of commands.

## Installation

To install `tfstate-transfer`, you need to have Go installed. 
Then, you can build and install the tool using the following command:

```bash
go install github.com/kassett/tfstate-transfer@latest
```

## Usage
There are two ways to use the ``tfstate-transfer`` program. The first
>>>>>>> 0.1.1-fixing-tag-logic
of these ways is via the simple CLI interface. 
When invoking tfstate-transfer, you pass a source directory and a target directory, as follows.

```shell
<<<<<<< HEAD
tfstate-transfer -sourceDir startDirectory -targetDir endDirectory -r module.db
```
=======
tfstate-transfer --source-dir startDirectory -target-dir endDirectory --r module.db --r module.elb
```
Optionally, the ``--dry-run`` flag can be passed, which will
simply print out commands instead of actually executing them.

It is also worth noting that if a resource has a different name
in the target directory, that can be specified by separating with a colon, as follows:
``--r module.db_source:module.db_target``

The other way to execute ``tfstate-transfer`` is to pass a JSON configuration file 
via the ``--config-dir`` argument.
This configuration file must contain a ``sourceDir``, a ``targetDir``,
and a list of ``resources``.
The list of resources is a map with a ``source`` resource and a ``target`` resource.
An example is shown below:
```json
{
  "sourceDir": "/Users/sourceDirectory",
  "targetDir": "/Users/targetDirectory",
  "resources": [
    {
      "source": "module.db_source",
      "target": "module.db_target"
    }
  ]
}
```
### Example Flow
1. Make a new Terraform environment
2. Copy the desired resources to the new .tf files. <b>DO NOT APPLY</b>
3. Using the CLI, run tfstate-transfer, specifying the specific resources to transfer.
4. Manually import any resources that failed automatic import.
