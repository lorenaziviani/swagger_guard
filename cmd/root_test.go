package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

var testRedisHost = "localhost"
var testRedisPort = "6379"
var testRedisClient *redis.Client
var testCtx = context.Background()

func setupTestRedis() {
	testRedisClient = redis.NewClient(&redis.Options{
		Addr: testRedisHost + ":" + testRedisPort,
		DB:   0,
	})
	testRedisClient.Del(testCtx, "metrics:executions", "metrics:high", "metrics:medium", "metrics:low", "metrics:last_run")
}

func teardownTestRedis() {
	testRedisClient.Del(testCtx, "metrics:executions", "metrics:high", "metrics:medium", "metrics:low", "metrics:last_run")
	testRedisClient.Close()
}

func Test_startsWithHTTPS(t *testing.T) {
	if !startsWithHTTPS("https://example.com") {
		t.Error("Deve retornar true para https://")
	}
	if startsWithHTTPS("http://example.com") {
		t.Error("Deve retornar false para http://")
	}
}

func Test_containsWord(t *testing.T) {
	if !containsWord("createUser", "create") {
		t.Error("Deve encontrar 'create' em 'createUser'")
	}
	if containsWord("deleteUser", "create") {
		t.Error("Não deve encontrar 'create' em 'deleteUser'")
	}
}

func Test_parseCmd_Integration(t *testing.T) {
	setupTestRedis()
	defer teardownTestRedis()

	os.Setenv("SWAGGER_GUARD_ALLOW_ABS_PATH", "1")
	defer os.Unsetenv("SWAGGER_GUARD_ALLOW_ABS_PATH")

	spec := `openapi: 3.0.0
info:
  title: Test API
  version: "1.0.0"
servers:
  - url: http://api.insecure.com
paths:
  /users:
    get:
      summary: List users
      security: []
      responses:
        '200':
          description: OK
`
	tempFile, err := os.CreateTemp("", "api-spec-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	if _, err := tempFile.Write([]byte(spec)); err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	exitCode, output := RunParse(tempFile.Name(), "cli", "", "", false)

	if !strings.Contains(output, "No Authentication") {
		t.Errorf("Expected alert for 'No Authentication', but not found. Output: %s", output)
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for critical failure, got: %d", exitCode)
	}
}
