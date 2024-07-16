package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Resource struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type ConfigFile struct {
	SourceDir string     `json:"SourceDir"`
	TargetDir string     `json:"TargetDir"`
	Resources []Resource `json:"resources"`
}

var (
	Resources      []string
	SourceDir      string
	TargetDir      string
	ConfigFileName string
	DryRun         bool
)

func UnmarshallConfigFileContent(configFileContent string) (string, string, []string, map[string]string) {
	var config ConfigFile
	err := json.Unmarshal([]byte(configFileContent), &config)
	if err != nil {
		Panic(fmt.Sprintf("The configuration file %s is not a valid JSON.", configFileContent))
	}

	var resourceList []string
	resourceMapping := make(map[string]string)

	for _, resource := range config.Resources {
		resourceList = append(resourceList, resource.Source)
		resourceMapping[resource.Source] = resource.Target
	}

	return config.SourceDir, config.TargetDir, resourceList, resourceMapping
}

func OpenConfigFile(configFilePath string) string {
	file, err := os.Open(configFilePath)
	if err != nil {
		Panic(fmt.Sprintf("The configuration file %s does not exist.", configFilePath))
	}

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		Panic(fmt.Sprintf("The configuration file %s could not be read..", configFilePath))
	}
	return string(byteValue)
}

func PullAliasesOutFromCli(resources []string) ([]string, map[string]string) {
	newResourceList := make([]string, 0)
	resourceMapping := make(map[string]string)

	for _, resource := range resources {
		if strings.Contains(resource, ":") {
			splitResource := strings.SplitN(resource, ":", 2)
			originalResource := splitResource[0]
			newResource := splitResource[1]
			resourceMapping[originalResource] = newResource
			newResourceList = append(newResourceList, originalResource)
		} else {
			newResourceList = append(newResourceList, resource)
			resourceMapping[resource] = resource
		}
	}

	return newResourceList, resourceMapping
}

func checkPath(dir string) string {
	_, err := os.Stat(dir)
	if err != nil {
		Panic(fmt.Sprintf("The directory %s does not exist.", dir))
	}
	path, _ := filepath.Abs(dir)
	return path
}

func ParseArguments() (string, string, map[string]string, bool) {
	var resourceMapping map[string]string

	if ConfigFileName != "" {
		configFileContent := OpenConfigFile(ConfigFileName)
		SourceDir, TargetDir, Resources, resourceMapping = UnmarshallConfigFileContent(configFileContent)
	} else {
		Resources, resourceMapping = PullAliasesOutFromCli(Resources)
	}

	if SourceDir == "" || TargetDir == "" {
		Panic("Both a source directory and a target directory must be specified.")
	}

	if len(Resources) == 0 {
		Panic("A list of resources must be specified, either via the configuration file or using --r.")
	}

	return SourceDir, TargetDir, resourceMapping, DryRun
}
