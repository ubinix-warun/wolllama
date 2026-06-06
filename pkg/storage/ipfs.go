package storage

import "fmt"

// IPFSProvider uploads blobs to IPFS via Pinata or another gateway.
// Not yet implemented.
type IPFSProvider struct{}

func NewIPFSProvider(cfg Config) (*IPFSProvider, error) {
	return nil, fmt.Errorf("ipfs provider not yet implemented")
}

func (i *IPFSProvider) Name() string       { return "ipfs" }
func (i *IPFSProvider) MaxChunkSize() int64 { return 100 * 1024 * 1024 } // 100 MB

func (i *IPFSProvider) Upload(data []byte) (string, error) {
	return "", fmt.Errorf("ipfs provider not yet implemented")
}
