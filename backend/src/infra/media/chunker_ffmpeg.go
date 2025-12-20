package media

import "podcast-summarizer/src/adapters/transcription"

type FFmpegChunker struct {
	Path string
}

func NewFFmpegChunker(path string) *FFmpegChunker {
	return &FFmpegChunker{Path: path}
}

func (c *FFmpegChunker) Chunk(inputPath string) ([]string, error) {
	return transcription.ChunkAudio(c.Path, inputPath)
}
