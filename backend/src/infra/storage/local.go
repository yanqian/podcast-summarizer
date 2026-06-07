package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalUploader struct {
	root string
}

func NewLocalUploader(root string) *LocalUploader {
	if root == "" {
		root = "data/storage"
	}
	return &LocalUploader{root: root}
}

func (u *LocalUploader) Upload(ctx context.Context, key string, localPath string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	dest := filepath.Join(u.root, filepath.Clean(key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("create local storage dir: %w", err)
	}

	src, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open upload source: %w", err)
	}
	defer src.Close()

	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create local storage object: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", fmt.Errorf("copy local storage object: %w", err)
	}
	return dest, nil
}

func BuildObjectKey(podcastID, filename string) string {
	return filepath.Join(podcastID, filename)
}

type NoopUploader struct{}

func (n *NoopUploader) Upload(ctx context.Context, key, path string) (string, error) {
	return "", nil
}
