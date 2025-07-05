/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

var filePath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "swagger_guard",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse an OpenAPI (Swagger) specification file",
	Run: func(cmd *cobra.Command, args []string) {
		if filePath == "" {
			fmt.Println("Please provide a file path with --file")
			os.Exit(1)
		}
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		var loader openapi3.Loader
		doc, err := loader.LoadFromData(data)
		if err != nil {
			fmt.Printf("Error parsing OpenAPI spec: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Paths:")
		for path, pathItem := range doc.Paths.Map() {
			fmt.Printf("- %s\n", path)
			for method := range pathItem.Operations() {
				fmt.Printf("  - Method: %s\n", method)
			}
		}
		fmt.Println("Security:", doc.Security)
		fmt.Println("Parameters:")
		for _, pathItem := range doc.Paths.Map() {
			for _, op := range pathItem.Operations() {
				for _, param := range op.Parameters {
					fmt.Printf("- %s (%s)\n", param.Value.Name, param.Value.In)
				}
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	parseCmd.Flags().StringVar(&filePath, "file", "", "Path to the OpenAPI spec file (.yaml or .json)")
	rootCmd.AddCommand(parseCmd)
}
