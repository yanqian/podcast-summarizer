package static_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultRuntimeHasNoNonSQLiteStoragePaths(t *testing.T) {
	repoRoot := findRepoRoot(t)
	checks := map[string][]string{
		filepath.Join(repoRoot, "backend", "cmd", "server", "main.go"): {
			"STORAGE_DRIVER",
			"OBJECT_STORAGE_DRIVER",
			"POSTGRES_URL",
			"VALKEY_URL",
			"R2_",
			"github.com/jackc/pgx",
			"github.com/redis/go-redis",
			"github.com/aws/aws-sdk-go-v2",
		},
		filepath.Join(repoRoot, "backend", "src", "config", "config.go"): {
			"STORAGE_DRIVER",
			"OBJECT_STORAGE_DRIVER",
			"POSTGRES_URL",
			"VALKEY_URL",
			"R2_",
		},
		filepath.Join(repoRoot, "backend", "src", "api", "router.go"): {
			"github.com/jackc/pgx",
			"github.com/redis/go-redis",
			"Postgres",
			"Valkey",
		},
		filepath.Join(repoRoot, "backend", "go.mod"): {
			"github.com/jackc/pgx",
			"github.com/redis/go-redis",
			"github.com/aws/aws-sdk-go-v2",
		},
		filepath.Join(repoRoot, "backend", "go.sum"): {
			"github.com/jackc/pgx",
			"github.com/redis/go-redis",
			"github.com/aws/aws-sdk-go-v2",
		},
	}

	for path, forbidden := range checks {
		assertFileDoesNotContain(t, path, forbidden)
	}
}

func TestDocsDoNotAdvertiseRemoteStorageSetup(t *testing.T) {
	repoRoot := findRepoRoot(t)
	forbidden := []string{
		"Postgres",
		"PostgreSQL",
		"Valkey",
		"Redis",
		"Cloudflare R2",
		"STORAGE_DRIVER",
		"OBJECT_STORAGE_DRIVER",
		"POSTGRES_URL",
		"VALKEY_URL",
		"R2_",
		"cloud mode",
	}
	paths := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "DEPLOYMENT.md"),
		filepath.Join(repoRoot, "backend", "README.md"),
		filepath.Join(repoRoot, "frontend", "README.md"),
		filepath.Join(repoRoot, "backend", ".env.example"),
		filepath.Join(repoRoot, "specs", "001-podcast-summary-ui", "quickstart.md"),
		filepath.Join(repoRoot, "specs", "001-podcast-summary-ui", "plan.md"),
		filepath.Join(repoRoot, "specs", "001-podcast-summary-ui", "research.md"),
		filepath.Join(repoRoot, "specs", "001-podcast-summary-ui", "data-model.md"),
		filepath.Join(repoRoot, "specs", "001-podcast-summary-ui", "tasks.md"),
	}

	for _, path := range paths {
		assertFileDoesNotContain(t, path, forbidden)
	}
}

func assertFileDoesNotContain(t *testing.T, path string, forbidden []string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(body)
	for _, term := range forbidden {
		if strings.Contains(text, term) {
			t.Fatalf("%s contains forbidden local-runtime term %q", path, term)
		}
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "backend", "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find repository root")
		}
		wd = parent
	}
}
