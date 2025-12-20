package transcription

import "errors"

// Client is a stub for transcription service integration.
type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Transcribe(url string) ([]byte, error) {
	if url == "" {
		return nil, errors.New("url required")
	}
	return []byte("Generated transcript paragraph"), nil
}
