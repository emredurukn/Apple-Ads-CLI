package appleads

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/emredurukan/asactl/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

func generateTestECDSAPEM(t *testing.T) ([]byte, *ecdsa.PrivateKey) {
	t.Helper()
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ecdsa key: %v", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("failed to marshal pkcs8: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	return pem.EncodeToMemory(pemBlock), privKey
}

func TestParsePrivateKey(t *testing.T) {
	pemBytes, expectedKey := generateTestECDSAPEM(t)

	parsedKey, err := ParsePrivateKey(pemBytes)
	if err != nil {
		t.Fatalf("ParsePrivateKey failed: %v", err)
	}

	if parsedKey.X.Cmp(expectedKey.X) != 0 || parsedKey.Y.Cmp(expectedKey.Y) != 0 {
		t.Errorf("parsed key public coordinates do not match expected")
	}
}

func TestGenerateClientSecret(t *testing.T) {
	pemBytes, _ := generateTestECDSAPEM(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "AuthKey_TEST123.p8")

	if err := os.WriteFile(keyPath, pemBytes, 0600); err != nil {
		t.Fatalf("failed to write temp key file: %v", err)
	}

	prof := &config.Profile{
		Name:           "test",
		KeyID:          "KEY123",
		TeamID:         "TEAM456",
		ClientID:       "SEARCHADS.client789",
		PrivateKeyPath: keyPath,
		BypassKeychain: true,
	}

	tokenString, err := GenerateClientSecret(prof)
	if err != nil {
		t.Fatalf("GenerateClientSecret failed: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected non-empty token string")
	}

	// Parse and verify token structure
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return public key for verification
		parsedKey, _ := ParsePrivateKey(pemBytes)
		return &parsedKey.PublicKey, nil
	})

	if err != nil {
		t.Fatalf("failed to parse/verify generated JWT: %v", err)
	}

	if !token.Valid {
		t.Fatal("token is not valid")
	}

	if token.Header["kid"] != "KEY123" {
		t.Errorf("expected kid KEY123, got %v", token.Header["kid"])
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected map claims")
	}

	if claims["sub"] != "SEARCHADS.client789" {
		t.Errorf("expected sub SEARCHADS.client789, got %v", claims["sub"])
	}
	if claims["iss"] != "TEAM456" {
		t.Errorf("expected iss TEAM456, got %v", claims["iss"])
	}
	if claims["aud"] != "https://appleid.apple.com" {
		t.Errorf("expected aud https://appleid.apple.com, got %v", claims["aud"])
	}
}
