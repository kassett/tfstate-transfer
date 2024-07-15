package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func unmarshallConfig(configFile string) (string, string, []string, map[string]string) {
	file, err := os.Open(configFile)
	if err != nil {
		os.Exit(1)
	}

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	var config ConfigFile
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("Error unmarshalling JSON file: ", err)
		os.Exit(1)
	}

	var resourceList []string
	resourceMapping := make(map[string]string)

	for _, resource := range config.Resources {
		resourceList = append(resourceList, resource.Source)
		resourceMapping[resource.Source] = resource.Target
	}

	return config.SourceDir, config.TargetDir, resourceList, resourceMapping
}

func pullAliasesOutFromCli(resources []string) ([]string, map[string]string) {
	newResourceList := make([]string, 0)
	resourceMapping := make(map[string]string)

	for _, resource := range resources {
		if strings.Contains(resource, ":") {
			splitResource := strings.SplitN(resource, ":", 2)
			originalResource := splitResource[0]
			newResource := splitResource[1]
			newResourceList = append(newResourceList, newResource)
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
		log.Fatal(fmt.Errorf("the file %s does not exist", dir))
	}
	path, _ := filepath.Abs(dir)
	return path
}

func ParseArguments() (string, string, map[string]string, bool) {
	var resourceMapping map[string]string

	if ConfigFileName != "" {
		SourceDir, TargetDir, Resources, resourceMapping = unmarshallConfig(ConfigFileName)
	} else {
		Resources, resourceMapping = pullAliasesOutFromCli(Resources)
	}

	if SourceDir == "" || TargetDir == "" {
		fmt.Println("Both a source directory and a target directory must be specified.")
	}

	if len(Resources) == 0 {
		fmt.Println("Resources must be specified.")
	}

	return SourceDir, TargetDir, resourceMapping, DryRun
}
