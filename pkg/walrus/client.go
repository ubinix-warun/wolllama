// Package walrus provides a thin convenience wrapper around
// github.com/namihq/walrus-go for wolllama-specific operations.
package walrus

import (
	"fmt"
	"io"

	walrusgo "github.com/namihq/walrus-go"
)

// Client wraps the walrus-go SDK client with wolllama-specific defaults.
type Client struct {
	inner *walrusgo.Client
}

// Config holds Walrus connection parameters.
type Config struct {
	PublisherURLs  []string
	AggregatorURLs []string
}

// NewClient creates a Client with the given config.
func NewClient(cfg Config) *Client {
	opts := []walrusgo.ClientOption{}
	if len(cfg.PublisherURLs) > 0 {
		opts = append(opts, walrusgo.WithPublisherURLs(cfg.PublisherURLs))
	}
	if len(cfg.AggregatorURLs) > 0 {
		opts = append(opts, walrusgo.WithAggregatorURLs(cfg.AggregatorURLs))
	}
	return &Client{inner: walrusgo.NewClient(opts...)}
}

// StoreBlob uploads data and returns the blob's object ID.
func (c *Client) StoreBlob(data []byte, epochs int) (string, error) {
	resp, err := c.inner.Store(data, &walrusgo.StoreOptions{Epochs: epochs})
	if err != nil {
		return "", fmt.Errorf("store blob: %w", err)
	}
	if resp.NewlyCreated != nil {
		return resp.NewlyCreated.BlobObject.BlobID, nil
	}
	if resp.AlreadyCertified != nil {
		return resp.AlreadyCertified.BlobID, nil
	}
	return "", fmt.Errorf("unexpected store response: no blob ID")
}

// StoreBlobFromReader uploads data from a reader and returns the blob's object ID.
func (c *Client) StoreBlobFromReader(r io.Reader, epochs int) (string, error) {
	resp, err := c.inner.StoreFromReader(r, &walrusgo.StoreOptions{Epochs: epochs})
	if err != nil {
		return "", fmt.Errorf("store blob: %w", err)
	}
	if resp.NewlyCreated != nil {
		return resp.NewlyCreated.BlobObject.BlobID, nil
	}
	if resp.AlreadyCertified != nil {
		return resp.AlreadyCertified.BlobID, nil
	}
	return "", fmt.Errorf("unexpected store response: no blob ID")
}

// StoreFile uploads a file and returns its blob object ID.
func (c *Client) StoreFile(path string, epochs int) (string, error) {
	resp, err := c.inner.StoreFile(path, &walrusgo.StoreOptions{Epochs: epochs})
	if err != nil {
		return "", fmt.Errorf("store file: %w", err)
	}
	if resp.NewlyCreated != nil {
		return resp.NewlyCreated.BlobObject.BlobID, nil
	}
	if resp.AlreadyCertified != nil {
		return resp.AlreadyCertified.BlobID, nil
	}
	return "", fmt.Errorf("unexpected store response: no blob ID")
}

// ReadBlob downloads a blob by its object ID.
func (c *Client) ReadBlob(blobID string) ([]byte, error) {
	data, err := c.inner.Read(blobID, nil)
	if err != nil {
		return nil, fmt.Errorf("read blob %s: %w", blobID, err)
	}
	return data, nil
}

// ReadBlobToFile downloads a blob and writes it to a local file.
func (c *Client) ReadBlobToFile(blobID, filePath string) error {
	if err := c.inner.ReadToFile(blobID, filePath, nil); err != nil {
		return fmt.Errorf("read blob %s to file %s: %w", blobID, filePath, err)
	}
	return nil
}

// HeadBlob returns metadata for a blob without downloading its content.
func (c *Client) HeadBlob(blobID string) (*walrusgo.BlobMetadata, error) {
	meta, err := c.inner.Head(blobID)
	if err != nil {
		return nil, fmt.Errorf("head blob %s: %w", blobID, err)
	}
	return meta, nil
}
