package storage

import (
	"fmt"

	wwalrus "github.com/wolllama/pkg/walrus"
)

const walrusMaxChunkSize = 256 * 1024 * 1024 // 256 MB

// WalrusProvider uploads blobs directly to a Walrus publisher.
type WalrusProvider struct {
	client *wwalrus.Client
	epochs int
}

// NewWalrusProvider creates a WalrusProvider from config.
func NewWalrusProvider(cfg Config) (*WalrusProvider, error) {
	client := wwalrus.NewClient(wwalrus.Config{
		PublisherURLs:  splitNonEmpty(cfg.PublisherURL),
		AggregatorURLs: splitNonEmpty(cfg.AggregatorURL),
	})
	epochs := cfg.Epochs
	if epochs <= 0 {
		epochs = 10
	}
	return &WalrusProvider{client: client, epochs: epochs}, nil
}

func (w *WalrusProvider) Name() string        { return "walrus" }
func (w *WalrusProvider) MaxChunkSize() int64  { return walrusMaxChunkSize }

func (w *WalrusProvider) Upload(data []byte) (string, error) {
	objID, err := w.client.StoreBlob(data, w.epochs)
	if err != nil {
		return "", fmt.Errorf("walrus upload: %w", err)
	}
	return objID, nil
}

func splitNonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}
