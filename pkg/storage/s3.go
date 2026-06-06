package storage

import "fmt"

// S3Provider uploads blobs to AWS S3.
// Not yet implemented.
type S3Provider struct{}

func NewS3Provider(cfg Config) (*S3Provider, error) {
	return nil, fmt.Errorf("s3 provider not yet implemented")
}

func (s *S3Provider) Name() string       { return "s3" }
func (s *S3Provider) MaxChunkSize() int64 { return 5 * 1024 * 1024 * 1024 } // 5 GB (S3 multipart)

func (s *S3Provider) Upload(data []byte) (string, error) {
	return "", fmt.Errorf("s3 provider not yet implemented")
}
