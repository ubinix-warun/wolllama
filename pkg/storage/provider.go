// Package storage provides a pluggable storage backend abstraction for wolllama.
// Supported providers: walrus (direct), tatum (managed gateway), ipfs (pinata), s3 (aws).
package storage

import "fmt"

// Provider uploads blobs to a decentralized or cloud storage backend.
type Provider interface {
	// Upload stores data and returns the storage-native object identifier.
	Upload(data []byte) (blobID string, err error)

	// MaxChunkSize returns the maximum upload size in bytes for a single call.
	// Blobs larger than this must be split into chunks before calling Upload.
	MaxChunkSize() int64

	// Name returns the provider identifier ("walrus", "tatum", "ipfs", "s3").
	Name() string
}

// Config holds provider-specific configuration.
type Config struct {
	// Walrus
	PublisherURL  string
	AggregatorURL string
	Epochs        int

	// Tatum
	TatumAPIKey string
	TatumAPIURL string // defaults to https://api.tatum.io

	// IPFS / Pinata (future)
	PinataAPIKey    string
	PinataSecretKey string

	// S3 (future)
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
}

// New creates a Provider based on the mode string.
func New(mode string, cfg Config) (Provider, error) {
	switch mode {
	case "walrus", "":
		return NewWalrusProvider(cfg)
	case "tatum":
		return NewTatumProvider(cfg)
	case "ipfs":
		return nil, fmt.Errorf("ipfs provider not yet implemented")
	case "s3":
		return nil, fmt.Errorf("s3 provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown storage provider %q (valid: walrus, tatum, ipfs, s3)", mode)
	}
}
