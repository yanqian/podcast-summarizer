package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadFile downloads a URL to a temp file and returns its path.
func DownloadFile(url string) (string, error) {
	tmpDir := os.TempDir()
	outPath := filepath.Join(tmpDir, fmt.Sprintf("podcast-%d.mp3", time.Now().UnixNano()))

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return outPath, nil
}
