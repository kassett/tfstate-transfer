package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
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
	SourceDir string     `json:"sourceDir"`
	TargetDir string     `json:"targetDir"`
	Resources []Resource `json:"resources"`
}

var (
	resources  []string
	sourceDir  string
	targetDir  string
	configFile string
	dryRun     bool
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

func parseArguments() (string, string, map[string]string, bool) {
	var resourceMapping map[string]string

	if configFile != "" {
		sourceDir, targetDir, resources, resourceMapping = unmarshallConfig(configFile)
	} else {
		resources, resourceMapping = pullAliasesOutFromCli(resources)
	}

	if sourceDir == "" || targetDir == "" {
		fmt.Println("Both a source directory and a target directory must be specified.")
	}

	if len(resources) == 0 {
		fmt.Println("Resources must be specified.")
	}

	return sourceDir, targetDir, resourceMapping, dryRun
}

var rootCmd = &cobra.Command{
	Use:   "tfstate-transfer",
	Short: "A simple CLI tool for transferring resources between Terraform environments.",
	Run: func(cmd *cobra.Command, args []string) {
		sourceDir, targetDir, resourceMapping, dryRun := parseArguments()
		fmt.Printf("Source Directory: %s\n", sourceDir)
		fmt.Printf("Target Directory: %s\n", targetDir)
		fmt.Printf("Resources: %v\n", resourceMapping)
		fmt.Printf("Dry Run: %v\n", dryRun)
		Run(sourceDir, targetDir, resourceMapping, dryRun)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&sourceDir, "source-dir", "", "Source directory")
	rootCmd.PersistentFlags().StringVar(&targetDir, "target-dir", "", "Target directory")
	rootCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "Path to the configuration file")
	rootCmd.PersistentFlags().StringArrayVar(&resources, "r", []string{}, "List of resources.")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Perform a dry run without making any changes")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
