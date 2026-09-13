package appleads

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/emredurukn/asactl/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AppleTokenURL = "https://appleid.apple.com/auth/token"
	Scope         = "searchadsorg"
)

// ParsePrivateKey extracts an ECDSA private key from a PEM encoded byte slice.
func ParsePrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block: no PEM data found")
	}

	// First attempt: PKCS#8
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		ecdsaKey, ok := parsedKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key in PKCS#8 block is not ECDSA")
		}
		return ecdsaKey, nil
	}

	// Fallback attempt: EC private key (SEC 1)
	ecKey, errEC := x509.ParseECPrivateKey(block.Bytes)
	if errEC == nil {
		return ecKey, nil
	}

	return nil, fmt.Errorf("failed to parse private key as PKCS#8 (%v) or EC (%v)", err, errEC)
}

// GenerateClientSecret creates a signed ES256 JWT to authenticate with Apple ID.
func GenerateClientSecret(prof *config.Profile) (string, error) {
	if prof.KeyID == "" {
		return "", errors.New("missing KeyID in profile")
	}
	if prof.ClientID == "" {
		return "", errors.New("missing ClientID in profile")
	}
	if prof.TeamID == "" {
		return "", errors.New("missing TeamID in profile")
	}

	var keyBytes []byte
	var err error

	// Try reading from keychain if configured and not bypassed
	if !prof.BypassKeychain {
		savedKey, kerr := config.GetPrivateKeyFromKeyring(prof.Name)
		if kerr == nil && savedKey != "" {
			keyBytes = []byte(savedKey)
		}
	}

	// Fallback to PrivateKeyPath if not found in keyring
	if len(keyBytes) == 0 {
		if prof.PrivateKeyPath == "" {
			return "", errors.New("private key path is not set and key was not found in keyring")
		}
		// Expand home directory if starts with ~
		path := prof.PrivateKeyPath
		if strings.HasPrefix(path, "~/") {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[2:])
		}
		keyBytes, err = os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read private key from %s: %w", path, err)
		}
	}

	ecdsaKey, err := ParsePrivateKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	now := time.Now()
	// Apple accepts expiration up to 180 days
	claims := jwt.MapClaims{
		"sub": prof.ClientID,
		"aud": "https://appleid.apple.com",
		"iss": prof.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(180 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = prof.KeyID

	signedString, err := token.SignedString(ecdsaKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign client secret JWT: %w", err)
	}

	return signedString, nil
}

// CachedToken represents token with expiration stored locally.
type CachedToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// TokenManager handles requesting and caching Apple Ads OAuth2 access tokens.
type TokenManager struct {
	profile    *config.Profile
	httpClient *http.Client
	mu         sync.Mutex
}

// NewTokenManager creates a token manager for a given profile.
func NewTokenManager(profile *config.Profile, httpClient *http.Client) *TokenManager {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &TokenManager{
		profile:    profile,
		httpClient: httpClient,
	}
}

func (tm *TokenManager) cacheFilePath() (string, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("token_%s.json", tm.profile.Name)
	return filepath.Join(dir, filename), nil
}

func (tm *TokenManager) readCachedToken() (*CachedToken, error) {
	path, err := tm.cacheFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cached CachedToken
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	// Ensure there is at least 5 minutes of validity left
	if time.Now().Add(5 * time.Minute).Before(cached.ExpiresAt) {
		return &cached, nil
	}

	return nil, errors.New("cached token expired")
}

func (tm *TokenManager) writeCachedToken(token string, expiresIn int) {
	path, err := tm.cacheFilePath()
	if err != nil {
		return
	}

	cached := CachedToken{
		AccessToken: token,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	if data, err := json.Marshal(cached); err == nil {
		_ = os.WriteFile(path, data, 0600)
	}
}

// GetAccessToken returns a valid access token, fetching a fresh one if necessary.
func (tm *TokenManager) GetAccessToken(ctx context.Context, forceRefresh bool) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !forceRefresh {
		if cached, err := tm.readCachedToken(); err == nil && cached != nil {
			return cached.AccessToken, nil
		}
	}

	clientSecret, err := GenerateClientSecret(tm.profile)
	if err != nil {
		return "", fmt.Errorf("failed to generate client secret: %w", err)
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", tm.profile.ClientID)
	form.Set("client_secret", clientSecret)
	form.Set("scope", Scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AppleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to build token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error during token exchange: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Apple ID token endpoint returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("empty access token in Apple response")
	}

	tm.writeCachedToken(tokenResp.AccessToken, tokenResp.ExpiresIn)
	return tokenResp.AccessToken, nil
}
