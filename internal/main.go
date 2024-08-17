package internal

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

type DryRunSet struct {
	deleteCommand  *string
	importCommands []string
}

func NewDryRunSet() *DryRunSet {
	return &DryRunSet{
		deleteCommand:  nil,
		importCommands: make([]string, 0),
	}
}

func (d *DryRunSet) SetDeleteCommand(command string) {
	d.deleteCommand = &command
}

func (d *DryRunSet) AddImportCommand(command string) {
	d.importCommands = append(d.importCommands, command)
}

func transferState(rn *RunHandler, sourceDir string, targetDir string, dryRunSet map[string]*DryRunSet) {

	for rn.HasNextResource() {
		resource, _ := rn.GetNextResource()
		command, err := runImport(targetDir, *resource, dryRunSet != nil)

		rn.ReportImportRun(resource.sourceName, resource.targetName, resource.topLevelName, err)

		if dryRunSet != nil {
			// Check if the key exists in the map; if not, create a new DryRunSet entry
			if _, exists := dryRunSet[resource.topLevelName]; !exists {
				fmt.Print(1)
			}
			// Add the import command to the DryRunSet
			dryRunSet[resource.topLevelName].AddImportCommand(command)
		}
	}

	resourcesToDelete := rn.ResourcesToDelete()
	for _, deleteResource := range resourcesToDelete {
		command := terraformRemoveState(deleteResource, sourceDir, dryRunSet != nil)
		if dryRunSet != nil {
			// Ensure the key exists before setting the delete command
			if _, exists := dryRunSet[deleteResource]; !exists {
				fmt.Println(1)
			}
			// Set the delete command
			dryRunSet[deleteResource].SetDeleteCommand(command)
		}
	}
}

func PrintDryRun(dryRunSet map[string]*DryRunSet) {
	for topLevelName, dryRun := range dryRunSet {

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetHeader([]string{fmt.Sprintf("Commands for transferring %s", topLevelName)})

		// Set table style options
		table.SetBorder(true)
		table.SetAutoWrapText(false)
		table.SetRowLine(true)
		table.SetColumnSeparator("│")

		// Add import commands with green text
		for _, command := range dryRun.importCommands {
			table.Rich(
				[]string{command},
				[]tablewriter.Colors{
					{tablewriter.FgGreenColor},
				},
			)
		}

		table.Rich(
			[]string{*dryRun.deleteCommand},
			[]tablewriter.Colors{
				{tablewriter.FgRedColor},
			},
		)

		table.Render()
	}
}

func Run(sourceDir string, targetDir string, resourceMapping map[string]string, dryRun bool) {
	sourceDir = checkPath(sourceDir)
	targetDir = checkPath(targetDir)

	stateFileContent := generateStateFile(sourceDir)
	runHandler := NewRunHandler(stateFileContent, resourceMapping)

	dryRunSet := map[string]*DryRunSet{}
	for topLevel, _ := range runHandler.topLevelResourceMapping {
		dryRunSet[topLevel] = NewDryRunSet()
	}

	transferState(runHandler, sourceDir, targetDir, dryRunSet)

	if !dryRun {
		runHandler.PrintFullRun()
	} else {
		PrintDryRun(dryRunSet)
	}

}
