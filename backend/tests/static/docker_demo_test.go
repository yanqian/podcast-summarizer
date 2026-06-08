package static_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerDemoIsLocalOnlyAndDocumented(t *testing.T) {
	repoRoot := findRepoRoot(t)
	composePath := filepath.Join(repoRoot, "compose.yaml")
	compose := readText(t, composePath)

	requiredComposeTerms := []string{
		"backend:",
		"frontend:",
		"./backend/data:/app/data",
		"SQLITE_PATH: /app/data/podcast.db",
		"LOCAL_STORAGE_PATH: /app/data/storage",
		"VITE_API_BASE_URL: http://localhost:8080",
		"condition: service_healthy",
	}
	for _, term := range requiredComposeTerms {
		if !strings.Contains(compose, term) {
			t.Fatalf("%s missing required local Docker demo term %q", composePath, term)
		}
	}

	forbidden := []string{
		"postgres",
		"redis",
		"valkey",
		"r2",
		"minio",
	}
	for _, term := range forbidden {
		if strings.Contains(strings.ToLower(compose), term) {
			t.Fatalf("%s contains non-local Docker demo dependency %q", composePath, term)
		}
	}

	verifyPath := filepath.Join(repoRoot, "scripts", "verify-docker-demo.sh")
	info, err := os.Stat(verifyPath)
	if err != nil {
		t.Fatalf("stat %s: %v", verifyPath, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("%s must be executable", verifyPath)
	}

	verify := readText(t, verifyPath)
	for _, term := range []string{"docker compose up --build -d", "http://127.0.0.1:8080/health", "http://127.0.0.1:8081/health"} {
		if !strings.Contains(verify, term) {
			t.Fatalf("%s missing verification term %q", verifyPath, term)
		}
	}

	readme := readText(t, filepath.Join(repoRoot, "README.md"))
	deployment := readText(t, filepath.Join(repoRoot, "DEPLOYMENT.md"))
	for _, body := range []struct {
		name string
		text string
	}{
		{name: "README.md", text: readme},
		{name: "DEPLOYMENT.md", text: deployment},
	} {
		for _, term := range []string{"./scripts/verify-docker-demo.sh", "http://localhost:8081", "backend/data"} {
			if !strings.Contains(body.text, term) {
				t.Fatalf("%s missing documented Docker demo term %q", body.name, term)
			}
		}
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}
