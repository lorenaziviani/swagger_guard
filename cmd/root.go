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

		// --- OWASP Top 10 Checks ---
		failures := make(map[string][]string)

		// 1. Routes without authentication (security: [])
		for path, pathItem := range doc.Paths.Map() {
			for method, op := range pathItem.Operations() {
				if op.Security != nil && len(*op.Security) == 0 {
					failures["No Authentication"] = append(failures["No Authentication"], fmt.Sprintf("%s %s", method, path))
				}
				// 2. GET for create/delete
				if method == "GET" && (op.OperationID != "" && (containsWord(op.OperationID, "create") || containsWord(op.OperationID, "delete"))) {
					failures["GET used for create/delete"] = append(failures["GET used for create/delete"], fmt.Sprintf("%s %s (operationId: %s)", method, path, op.OperationID))
				}
				// 5. Insecure HTTP Methods
				if method == "TRACE" || method == "OPTIONS" {
					failures["Insecure HTTP Methods"] = append(failures["Insecure HTTP Methods"], fmt.Sprintf("%s %s", method, path))
				}
				// 4. Query parameters without type
				for _, param := range op.Parameters {
					if param.Value.In == "query" && param.Value.Schema == nil {
						failures["Query parameter without type"] = append(failures["Query parameter without type"], fmt.Sprintf("%s %s param: %s", method, path, param.Value.Name))
					}
				}
			}
		}
		// 3. Absence of HTTPS
		for _, server := range doc.Servers {
			if server.URL != "" && !startsWithHTTPS(server.URL) {
				failures["No HTTPS"] = append(failures["No HTTPS"], server.URL)
			}
		}

		if len(failures) == 0 {
			fmt.Println("No OWASP Top 10 issues found!")
		} else {
			fmt.Println("OWASP Top 10 Issues:")
			for category, items := range failures {
				fmt.Printf("\n[%s]\n", category)
				for _, item := range items {
					fmt.Printf("- %s\n", item)
				}
			}
		}

		fmt.Println("\nPaths:")
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

func startsWithHTTPS(url string) bool {
	return len(url) >= 8 && url[:8] == "https://"
}

func containsWord(s, word string) bool {
	return len(s) > 0 && len(word) > 0 && (containsIgnoreCase(s, word))
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (stringContainsFold(s, substr))
}

func stringContainsFold(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len([]rune(s)) >= len([]rune(substr)) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || stringContainsFold(s[1:], substr)))) || (len(s) > len(substr) && stringContainsFold(s[1:], substr)))
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
