package main

func transferState(rn *RunHandler, sourceDir string, targetDir string, dryRun bool) {
	for rn.HasNextResource() {
		resource, _ := rn.GetNextResource()
		err := runImport(targetDir,
			*resource)
		rn.ReportImportRun(resource.sourceName, resource.targetName, resource.topLevelName, err)
	}

	resourcesToDelete := rn.ResourcesToDelete()
	for _, deleteResource := range resourcesToDelete {
		terraformRemoveState(deleteResource, sourceDir)
	}
}

func Run(sourceDir string, targetDir string, resourceMapping map[string]string, dryRun bool) {
	sourceDir = checkPath(sourceDir)
	targetDir = checkPath(targetDir)

	stateFileContent := generateStateFile(sourceDir)
	runHandler := NewRunHandler(stateFileContent, resourceMapping)

	transferState(runHandler, sourceDir, targetDir, dryRun)
	runHandler.FinishRun()
}
