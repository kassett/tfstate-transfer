package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/kassett/tfstate-transfer/internal"
)

var rootCmd = &cobra.Command{
	Use:   "tfstate-transfer",
	Short: "A simple CLI tool for transferring resources between Terraform environments.",
	Run: func(cmd *cobra.Command, args []string) {
		sourceDir, targetDir, resourceMapping, dryRun := internal.ParseArguments()
		internal.Run(sourceDir, targetDir, resourceMapping, dryRun)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&internal.SourceDir, "source-dir", "", "Source directory")
	rootCmd.PersistentFlags().StringVar(&internal.TargetDir, "target-dir", "", "Target directory")
	rootCmd.PersistentFlags().StringVar(&internal.ConfigFileName, "config-file", "", "Path to the configuration file")
	rootCmd.PersistentFlags().StringArrayVar(&internal.Resources, "r", []string{}, "List of resources.")
	rootCmd.PersistentFlags().BoolVar(&internal.DryRun, "dry-run", false, "Perform a dry run without making any changes")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
