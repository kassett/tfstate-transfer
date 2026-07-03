package internal

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
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
	dryRun := dryRunSet != nil

	for rn.HasNextResource() {
		resource, _ := rn.GetNextResource()
		command, err := runImport(targetDir, *resource, dryRun)

		rn.ReportImportRun(resource.sourceName, resource.targetName, resource.topLevelName, err)

		if dryRun {
			dryRunSet[resource.topLevelName].AddImportCommand(command)
		}
	}

	resourcesToDelete := rn.ResourcesToDelete()
	for _, deleteResource := range resourcesToDelete {
		command := terraformRemoveState(deleteResource, sourceDir, dryRun)
		if dryRun {
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

	var dryRunSet map[string]*DryRunSet
	if dryRun {
		dryRunSet = map[string]*DryRunSet{}
		for topLevel := range runHandler.topLevelResourceMapping {
			dryRunSet[topLevel] = NewDryRunSet()
		}
	}

	transferState(runHandler, sourceDir, targetDir, dryRunSet)

	if !dryRun {
		runHandler.PrintFullRun()
	} else {
		PrintDryRun(dryRunSet)
	}

}
