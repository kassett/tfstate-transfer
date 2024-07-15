//go:build tests
// +build tests

package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var requiredTestEnvVars = map[string]string{
	"AWS_ACCESS_KEY_ID":     "localstack",
	"AWS_SECRET_ACCESS_KEY": "localstack",
	"AWS_ENDPOINT_URL":      "http://localhost:4566/",
}

var temporaryDirectories = make([]string, 0)
var currentDirectory, _ = os.Getwd()

func setup() {
	// Setup environment variables for Localstack
	for ev, value := range requiredTestEnvVars {
		_ = os.Setenv(ev, value)
	}
}

func resetLocalstackState() {
	// JSON data to be sent in the POST request
	jsonData := []byte(`{"action":"restart"}`)

	// Create a new POST request
	req, err := http.NewRequest("POST",
		"http://localhost:4566/_localstack/health",
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the appropriate header for the request
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the default HTTP client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status: %s", resp.Status)
	}

	log.Println("Request succeeded.")
}

func teardown() {
	// Delete temporary directories
	for _, tempDir := range temporaryDirectories {
		_ = os.RemoveAll(tempDir)
	}

	_ = os.Chdir(currentDirectory)

	resetLocalstackState()
}

func setupTerraformTest(baseDirName string) string {
	currentDir, _ := filepath.Abs(func() string { _, file, _, _ := runtime.Caller(0); return file }())
	integrationTestDir := filepath.Dir(filepath.Dir(currentDir))
	testDir := filepath.Join(filepath.Join(integrationTestDir, "integration_tests"), baseDirName)

	tempDir, err := os.MkdirTemp("/tmp", "tempDir-")
	if err != nil {
		fmt.Printf("Failed to create temporary directory %s\n", tempDir)
	}

	// Copy the source directory to the temporary directory
	err = filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the corresponding path in the destination directory
		relPath, err := filepath.Rel(testDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(tempDir, relPath)

		// If it's a directory, create it in the destination directory
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// If it's a file, copy it to the destination directory
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(sourceFile *os.File) {
			_ = sourceFile.Close()
		}(sourceFile)

		destinationFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer func(destinationFile *os.File) {
			_ = destinationFile.Close()
		}(destinationFile)

		if _, err := io.Copy(destinationFile, sourceFile); err != nil {
			return err
		}

		// Ensure the copied file has the same permissions as the original file
		return os.Chmod(destPath, info.Mode())
	})

	temporaryDirectories = append(temporaryDirectories, tempDir)
	return tempDir
}

func terraformPlanForCase(tempDir string) (string, string) {
	outputs := make(map[string]string)
	stateDirectories := []string{filepath.Join(tempDir, "source"), filepath.Join(tempDir, "target")}
	for _, dir := range stateDirectories {
		cmd := exec.Command("terraform", "plan")
		cmd.Dir = dir
		output, _ := cmd.CombinedOutput()

		if strings.HasSuffix(dir, "source") {
			outputs["source"] = string(output)
		} else {
			outputs["target"] = string(output)
		}
	}
	return outputs["source"], outputs["target"]
}

func initializeTerraformTest(tempDir string) error {
	stateDirectories := []string{filepath.Join(tempDir, "source"), filepath.Join(tempDir, "target")}
	for i, dir := range stateDirectories {
		cmd := exec.Command("terraform", "init")
		cmd.Dir = dir
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Failed to initialize the plan.")
			return err
		}

		if i == 0 {
			cmd = exec.Command("terraform", "apply", "-auto-approve")
			cmd.Dir = dir
			_, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Println("Failed to apply the Resources...")
				return err
			}
		}
	}

	_ = os.Chdir(tempDir)
	return nil
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestCase1(t *testing.T) {
	// In Case1, we want to test a simple transfer of 2 Resources
	// There is no fancy logic, just 2 of three Resources transferred
	// from one environment to another
	tempDir := setupTerraformTest("case1")
	err := initializeTerraformTest(tempDir)
	if err != nil {
		t.Error("Failed to setup the test...", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	sourceDir, targetDir, _, resourceMapping := unmarshallConfig(configFile)
	Run(sourceDir, targetDir, resourceMapping, false)

	sourcePlan, targetPlan := terraformPlanForCase(tempDir)

	assert.Contains(t, sourcePlan,
		"3 to add", "Expected the source plan to want to add 3 Resources")
	assert.Contains(t, targetPlan,
		"0 to add", "Expected the target plan to want to add 0 Resources")
}

func TestCase2(t *testing.T) {
	// In Case2, we want to test transferring Resources with multiple
	// instances. This can either be a simple count or a for_each loop

	// Additionally, we do a sneaky thing and check that
	// we can delete just one instance of a resource with a count

	tempDir := setupTerraformTest("case2")
	err := initializeTerraformTest(tempDir)
	if err != nil {
		t.Error("Failed to setup the test...", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	sourceDir, targetDir, _, resourceMapping := unmarshallConfig(configFile)
	Run(sourceDir, targetDir, resourceMapping, false)

	sourcePlan, targetPlan := terraformPlanForCase(tempDir)

	assert.Contains(t, sourcePlan,
		"8 to add", "Expected the source plan to want to add 8 Resources")
	assert.Contains(t, targetPlan,
		"3 to add", "Expected the target plan to want to add 3 Resources")
}

func TestCase3(t *testing.T) {
	// In Case3, we test transferring the state of a module
	// with multiple instances

	tempDir := setupTerraformTest("case3")
	err := initializeTerraformTest(tempDir)
	if err != nil {
		t.Error("Failed to setup the test...", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	sourceDir, targetDir, _, resourceMapping := unmarshallConfig(configFile)
	Run(sourceDir, targetDir, resourceMapping, false)

	sourcePlan, targetPlan := terraformPlanForCase(tempDir)

	assert.Contains(t, sourcePlan,
		"4 to add", "Expected the source plan to want to add 4 Resources")
	assert.Contains(t, targetPlan,
		"2 to add", "Expected the target plan to want to add 2 Resources")
}

func TestCase4(t *testing.T) {
	// In Case4, we test that we fail gracefully when import is not supported

	tempDir := setupTerraformTest("case4")
	err := initializeTerraformTest(tempDir)
	if err != nil {
		t.Error("Failed to setup the test...", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	sourceDir, targetDir, _, resourceMapping := unmarshallConfig(configFile)
	Run(sourceDir, targetDir, resourceMapping, false)

	sourcePlan, targetPlan := terraformPlanForCase(tempDir)

	assert.Contains(t, sourcePlan,
		"no changes are needed", "Expected the state not to change.")
	assert.Contains(t, targetPlan,
		"1 to add", "Expected the target plan to want to add 1 Resources")
}

func TestCase5(t *testing.T) {
	// In Case5, we test that we can import
	//based on a field other than the first one

	tempDir := setupTerraformTest("case5")
	err := initializeTerraformTest(tempDir)
	if err != nil {
		t.Error("Failed to setup the test...", err)
	}

	configFile := filepath.Join(tempDir, "config.json")
	sourceDir, targetDir, _, resourceMapping := unmarshallConfig(configFile)
	Run(sourceDir, targetDir, resourceMapping, false)

	sourcePlan, targetPlan := terraformPlanForCase(tempDir)

	assert.Contains(t, sourcePlan,
		"1 to add", "Expected the source plan to want to add 1 Resources")
	assert.Contains(t, targetPlan,
		"no changes are needed", "Expected the state not to change.")
}
