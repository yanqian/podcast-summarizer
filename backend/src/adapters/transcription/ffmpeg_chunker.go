package transcription

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// ChunkAudio uses ffmpeg to split an audio file into ~5 minute chunks.
func ChunkAudio(ffmpegPath, inputPath string) ([]string, error) {
	outputDir, err := os.MkdirTemp("", "podcast-chunks-*")
	if err != nil {
		return nil, fmt.Errorf("create chunk temp dir: %w", err)
	}
	outputPattern := filepath.Join(outputDir, "chunk-%03d.mp3")
	cmd := exec.Command(ffmpegPath, "-i", inputPath, "-f", "segment", "-segment_time", "300", "-c", "copy", outputPattern)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg chunking failed: %w", err)
	}

	matches, err := filepath.Glob(filepath.Join(outputDir, "chunk-*.mp3"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return nil, fmt.Errorf("ffmpeg produced no audio chunks")
	}
	return matches, nil
}
