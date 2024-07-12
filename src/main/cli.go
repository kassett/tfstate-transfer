package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Resource struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type ConfigFile struct {
	SourceDir string `json:"sourceDir"`
	TargetDir string `json:"targetDir"`

	Resources []Resource `json:"resources"`
}

type targetResources []string

func (i *targetResources) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *targetResources) String() string {
	return "resource"
}

var resources targetResources

func unmarshallConfig(configFile string) (*string, *string, []string, map[string]string) {
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

	return &config.SourceDir, &config.TargetDir, resourceList, resourceMapping
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

func parseArguments() (string, string, map[string]string, bool) {
	sourceDir := flag.String("source-dir", "", "Source directory")
	targetDir := flag.String("target-dir", "", "Target directory")
	configFile := flag.String("config-file", "", "Path to the configuration file")
	flag.Var(&resources, "r", "List of resources.")
	dryRun := flag.Bool("dry-run", false, "Perform a dry run without making any changes")
	flag.Parse()

	var resourceMapping map[string]string

	if *configFile != "" {
		sourceDir, targetDir, resources, resourceMapping = unmarshallConfig(*configFile)
	} else {
		resources, resourceMapping = pullAliasesOutFromCli(resources)
	}

	if sourceDir == nil || targetDir == nil {
		fmt.Println("Both a source directory and a target directory must be specified.")
	}

	if len(resources) == 0 {
		fmt.Println("Resources must be specified.")
	}

	return *sourceDir, *targetDir, resourceMapping, *dryRun
}
