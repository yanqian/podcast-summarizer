package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalUploaderCopiesFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	uploader := NewLocalUploader(filepath.Join(dir, "objects"))
	dest, err := uploader.Upload(context.Background(), BuildObjectKey("podcast-1", "chunk.txt"), src)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("unexpected copied content: %q", got)
	}
}
