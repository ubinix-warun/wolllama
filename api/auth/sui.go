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

// VerifySignature checks that a message was signed by the owner of a Sui wallet
// using signPersonalMessage. Reconstructs the IntentMessage wrapper that Sui
// wallets apply before signing.
func VerifySignature(publicKeyB64, signatureB64 string, message []byte) error {
	return VerifySignatureRaw(publicKeyB64, signatureB64, wrapIntentMessage(message))
}

// VerifySignatureRaw verifies a raw ed25519 signature without IntentMessage wrapping.
func VerifySignatureRaw(publicKeyB64, signatureB64 string, rawMessage []byte) error {
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

	if !ed25519.Verify(pubKey, rawMessage, sig) {
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

// wrapIntentMessage wraps raw bytes in Sui's IntentMessage BCS envelope.
//
// signPersonalMessage signs:
//   BCS(IntentMessage {
//       intent: Intent { scope: 3, version: 0, app_id: 0 },
//       value: BCS(vector<u8>) of raw message
//   })
//
// BCS layout:
//   [scope:u8, version:u8, app_id:u8]        — Intent (3 bytes)
//   ULEB128(len(value))                       — BCS length of value field (outer)
//   ULEB128(len(raw))                         — BCS length of message vector (inner)
//   raw bytes
func wrapIntentMessage(data []byte) []byte {
	intent := []byte{0x03, 0x00, 0x00}

	// Inner: BCS serialization of message as vector<u8>
	innerLen := uleb128(uint64(len(data)))
	inner := append(innerLen, data...)

	// Outer: BCS serialization of value field in IntentMessage (vector<u8>)
	outerLen := uleb128(uint64(len(inner)))

	result := make([]byte, 0, len(intent)+len(outerLen)+len(inner))
	result = append(result, intent...)
	result = append(result, outerLen...)
	result = append(result, inner...)
	return result
}

func uleb128(n uint64) []byte {
	var buf []byte
	for {
		b := byte(n & 0x7f)
		n >>= 7
		if n > 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if n == 0 {
			break
		}
	}
	return buf
}
