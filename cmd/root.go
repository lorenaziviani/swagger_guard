/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"

	"errors"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var filePath string
var outputFormat string
var outputFile string
var metricsDBPath string
var showMetrics bool
var ctx = context.Background()

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

// RunParse is the main function that parses the OpenAPI specification and checks for OWASP Top 10 issues
func RunParse(filePath, outputFormat, outputFile, metricsDBPath string, showMetrics bool) (int, string) {
	if showMetrics {
		client := openRedisClient()
		var sb strings.Builder
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		_ = printMetricsRedis(client)
		if err := w.Close(); err != nil {
			sb.WriteString("[WARN] error closing pipe: " + err.Error() + "\n")
		}
		os.Stdout = old
		out, _ := io.ReadAll(r)
		sb.Write(out)
		return 0, sb.String()
	}
	if filePath == "" {
		return 1, "Please provide a file path with --file"
	}
	if err := isSafePath(filePath); err != nil {
		return 1, "Invalid file path: " + err.Error()
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 1, fmt.Sprintf("Error reading file: %v", err)
	}
	var loader openapi3.Loader
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return 1, fmt.Sprintf("Error parsing OpenAPI spec: %v", err)
	}

	failures := make(map[string][]string)
	severity := map[string]string{
		"No Authentication":            "high",
		"Insecure HTTP Methods":        "high",
		"No HTTPS":                     "high",
		"GET used for create/delete":   "medium",
		"Query parameter without type": "low",
	}

	for path, pathItem := range doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			if op.Security != nil && len(*op.Security) == 0 {
				failures["No Authentication"] = append(failures["No Authentication"], fmt.Sprintf("%s %s", method, path))
			}
			if method == "GET" && (op.OperationID != "" && (containsWord(op.OperationID, "create") || containsWord(op.OperationID, "delete"))) {
				failures["GET used for create/delete"] = append(failures["GET used for create/delete"], fmt.Sprintf("%s %s (operationId: %s)", method, path, op.OperationID))
			}
			if method == "TRACE" || method == "OPTIONS" {
				failures["Insecure HTTP Methods"] = append(failures["Insecure HTTP Methods"], fmt.Sprintf("%s %s", method, path))
			}
			for _, param := range op.Parameters {
				if param.Value.In == "query" && param.Value.Schema == nil {
					failures["Query parameter without type"] = append(failures["Query parameter without type"], fmt.Sprintf("%s %s param: %s", method, path, param.Value.Name))
				}
			}
		}
	}
	for _, server := range doc.Servers {
		if server.URL != "" && !startsWithHTTPS(server.URL) {
			failures["No HTTPS"] = append(failures["No HTTPS"], server.URL)
		}
	}

	hasHigh := false
	highCount, mediumCount, lowCount := 0, 0, 0
	for category, items := range failures {
		sev := severity[category]
		if sev == "high" && len(items) > 0 {
			hasHigh = true
		}
		switch sev {
		case "high":
			highCount += len(items)
		case "medium":
			mediumCount += len(items)
		case "low":
			lowCount += len(items)
		}
	}

	client := openRedisClient()
	_ = updateMetricsRedis(client, highCount, mediumCount, lowCount)
	var sb strings.Builder
	if outputFormat == "json" {
		output := map[string]interface{}{"issues": []map[string]interface{}{}, "summary": map[string]int{"high": 0, "medium": 0, "low": 0}}
		for category, items := range failures {
			sev := severity[category]
			for _, item := range items {
				output["issues"] = append(output["issues"].([]map[string]interface{}), map[string]interface{}{"category": category, "severity": sev, "item": item})
				output["summary"].(map[string]int)[sev]++
			}
		}
		jsonBytes, _ := json.MarshalIndent(output, "", "  ")
		if outputFile != "" {
			if err := os.WriteFile(outputFile, jsonBytes, 0600); err != nil {
				return 1, "Error writing output file: " + err.Error()
			}
		}
		sb.WriteString(string(jsonBytes))
		if hasHigh {
			return 1, sb.String()
		}
		return 0, sb.String()
	}
	if outputFormat == "markdown" {
		sb.WriteString("# OWASP Top 10 Issues\n\n")
		for category, items := range failures {
			sev := severity[category]
			sb.WriteString(fmt.Sprintf("## %s (%s)\n", category, strings.ToUpper(sev)))
			for _, item := range items {
				sb.WriteString(fmt.Sprintf("- %s\n", item))
			}
			sb.WriteString("\n")
		}
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(sb.String()), 0600); err != nil {
				return 1, "Error writing output file: " + err.Error()
			}
		}
		if hasHigh {
			return 1, sb.String()
		}
		return 0, sb.String()
	}
	if len(failures) == 0 {
		if _, err := color.New(color.FgGreen).Fprintln(&sb, "No OWASP Top 10 issues found!"); err != nil {
			sb.WriteString("[WARN] error writing to buffer: " + err.Error() + "\n")
		}
	} else {
		if _, err := color.New(color.FgRed, color.Bold).Fprintln(&sb, "OWASP Top 10 Issues:"); err != nil {
			sb.WriteString("[WARN] error writing to buffer: " + err.Error() + "\n")
		}
		for category, items := range failures {
			sev := severity[category]
			var c *color.Color
			switch sev {
			case "high":
				c = color.New(color.FgRed, color.Bold)
			case "medium":
				c = color.New(color.FgYellow, color.Bold)
			case "low":
				c = color.New(color.FgYellow)
			default:
				c = color.New(color.FgWhite)
			}
			if _, err := c.Fprintf(&sb, "\n[%s] (%s)\n", category, strings.ToUpper(sev)); err != nil {
				sb.WriteString("[WARN] error writing to buffer: " + err.Error() + "\n")
			}
			for _, item := range items {
				if _, err := c.Fprintf(&sb, "- %s\n", item); err != nil {
					sb.WriteString("[WARN] error writing to buffer: " + err.Error() + "\n")
				}
			}
		}
		if outputFile != "" {
			var fileSB strings.Builder
			fileSB.WriteString("# OWASP Top 10 Issues\n\n")
			for category, items := range failures {
				sev := severity[category]
				fileSB.WriteString(fmt.Sprintf("## %s (%s)\n", category, strings.ToUpper(sev)))
				for _, item := range items {
					fileSB.WriteString(fmt.Sprintf("- %s\n", item))
				}
				fileSB.WriteString("\n")
			}
			if err := os.WriteFile(outputFile, []byte(fileSB.String()), 0600); err != nil {
				return 1, "Error writing output file: " + err.Error()
			}
		}
	}
	sb.WriteString(fmt.Sprintf("\nFound issues: high=%d, medium=%d, low=%d\n", highCount, mediumCount, lowCount))
	if hasHigh {
		return 1, sb.String()
	}
	return 0, sb.String()
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse an OpenAPI (Swagger) specification file",
	Run: func(cmd *cobra.Command, args []string) {
		exitCode, output := RunParse(filePath, outputFormat, outputFile, metricsDBPath, showMetrics)
		fmt.Print(output)
		os.Exit(exitCode)
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
	parseCmd.Flags().StringVar(&outputFormat, "output", "cli", "Output format: cli, json, markdown")
	parseCmd.Flags().StringVar(&outputFile, "output-file", "", "Output file path (optional)")
	parseCmd.Flags().StringVar(&metricsDBPath, "metrics-db", "metrics.db", "Path to metrics SQLite database")
	parseCmd.Flags().BoolVar(&showMetrics, "metrics", false, "Show accumulated metrics")
	rootCmd.AddCommand(parseCmd)
}

func openRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	addr := host + ":" + port
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
}

func updateMetricsRedis(client *redis.Client, high, medium, low int) error {
	pipe := client.TxPipeline()
	pipe.Incr(ctx, "metrics:executions")
	pipe.IncrBy(ctx, "metrics:high", int64(high))
	pipe.IncrBy(ctx, "metrics:medium", int64(medium))
	pipe.IncrBy(ctx, "metrics:low", int64(low))
	pipe.Set(ctx, "metrics:last_run", time.Now().Format(time.RFC3339), 0)
	_, err := pipe.Exec(ctx)
	return err
}

func printMetricsRedis(client *redis.Client) error {
	execs, _ := client.Get(ctx, "metrics:executions").Result()
	high, _ := client.Get(ctx, "metrics:high").Result()
	medium, _ := client.Get(ctx, "metrics:medium").Result()
	low, _ := client.Get(ctx, "metrics:low").Result()
	lastRun, _ := client.Get(ctx, "metrics:last_run").Result()
	fmt.Println("\n==== CLI Metrics (Redis) ====")
	fmt.Println("Total executions:", execs)
	fmt.Println("Total high severity issues:", high)
	fmt.Println("Total medium severity issues:", medium)
	fmt.Println("Total low severity issues:", low)
	fmt.Println("Last run:", lastRun)
	return nil
}

func isSafePath(filePath string) error {
	if os.Getenv("SWAGGER_GUARD_ALLOW_ABS_PATH") == "1" {
		return nil
	}
	if filepath.IsAbs(filePath) {
		return errors.New("absolute paths are not allowed")
	}
	if strings.Contains(filePath, "..") {
		return errors.New("path traversal is not allowed")
	}
	return nil
}
