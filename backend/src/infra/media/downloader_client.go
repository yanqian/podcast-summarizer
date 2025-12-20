package media

// FileDownloader wraps DownloadFile into an interface implementation.
type FileDownloader struct{}

func NewFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

func (d *FileDownloader) Download(url string) (string, error) {
	return DownloadFile(url)
}
