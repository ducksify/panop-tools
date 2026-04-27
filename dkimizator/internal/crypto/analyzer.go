package crypto

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/ducksify/panop-tools/dkimizator/internal/dkim"
)

// KeyInfo contains analyzed RSA key information
type KeyInfo struct {
	Modulus     *big.Int
	Exponent    *big.Int
	Size        int
	Fingerprint string
	Mode        string
}

// AnalyzeKey analyzes a DKIM record and extracts RSA key information
func AnalyzeKey(record *dkim.Record) (*KeyInfo, error) {
	if record.PublicKey == "" {
		return nil, fmt.Errorf("no public key in record")
	}

	// Decode base64 public key
	keyBytes, err := record.DecodePublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Parse DER-encoded public key
	pubKey, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	// Calculate fingerprint (SHA1 of base64-decoded key)
	hash := sha1.Sum(keyBytes)
	fingerprint := fmt.Sprintf("%x", hash)

	// Determine mode
	mode := "PROD"
	if record.TestMode {
		mode = "TEST"
	}

	return &KeyInfo{
		Modulus:     rsaPubKey.N,
		Exponent:    big.NewInt(int64(rsaPubKey.E)),
		Size:        rsaPubKey.Size() * 8,
		Fingerprint: fingerprint,
		Mode:        mode,
	}, nil
}

// FormatX509 formats the public key as X.509 PEM from DER bytes
func FormatX509(keyBytes []byte) string {
	encoded := base64.StdEncoding.EncodeToString(keyBytes)

	// Split into 64-character lines
	var lines []string
	for i := 0; i < len(encoded); i += 64 {
		end := i + 64
		if end > len(encoded) {
			end = len(encoded)
		}
		lines = append(lines, encoded[i:end])
	}

	return "-----BEGIN PUBLIC KEY-----\n" +
		strings.Join(lines, "\n") + "\n" +
		"-----END PUBLIC KEY-----\n"
}

// GetKeyBytes returns the DER-encoded public key bytes for a record
func GetKeyBytes(record *dkim.Record) ([]byte, error) {
	return record.DecodePublicKey()
}
