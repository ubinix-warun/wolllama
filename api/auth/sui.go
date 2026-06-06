// Package auth handles Sui wallet authentication via pure signature verification.
// No external API calls — verification uses standard ed25519/secp256k1 crypto.
package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// Nonce generates a random 32-byte challenge for wallet signing.
func Nonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// VerifySignature checks that a message was signed by the owner of a Sui wallet.
// publicKey is base64-encoded (ed25519 32-byte key).
// signature is base64-encoded (ed25519 64-byte signature).
// message is the raw bytes that were signed.
func VerifySignature(publicKeyB64, signatureB64 string, message []byte) error {
	pubKey, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}

	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	if len(pubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid ed25519 public key size: %d (expected %d)", len(pubKey), ed25519.PublicKeySize)
	}

	if !ed25519.Verify(pubKey, message, sig) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// ShortAddress returns a shortened display version of a Sui address.
func ShortAddress(addr string) string {
	addr = strings.TrimPrefix(addr, "0x")
	if len(addr) > 12 {
		return "0x" + addr[:6] + "..." + addr[len(addr)-4:]
	}
	return "0x" + addr
}

// ValidateAddress checks that a Sui address looks valid (0x + 64 hex chars).
func ValidateAddress(addr string) bool {
	addr = strings.TrimPrefix(addr, "0x")
	if len(addr) != 64 {
		return false
	}
	for _, c := range addr {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
