package transcription

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ChunkAudio uses ffmpeg to split an audio file into ~5 minute chunks.
func ChunkAudio(ffmpegPath, inputPath string) ([]string, error) {
	outputPattern := filepath.Join(os.TempDir(), "chunk-%03d.mp3")
	cmd := exec.Command(ffmpegPath, "-i", inputPath, "-f", "segment", "-segment_time", "300", "-c", "copy", outputPattern)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg chunking failed: %w", err)
	}

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(outputPattern), "chunk-*.mp3"))
	if err != nil {
		return nil, err
	}
	return matches, nil
}
