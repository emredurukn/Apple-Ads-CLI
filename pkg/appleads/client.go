package appleads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emredurukan/asactl/pkg/config"
)

const (
	DefaultBaseURL = "https://api.searchads.apple.com/api/v5"
)

// Client handles interaction with the Apple Search Ads API.
type Client struct {
	BaseURL      string
	Profile      *config.Profile
	TokenManager *TokenManager
	HTTPClient   *http.Client
}

// NewClient constructs an API client from a configuration profile.
func NewClient(prof *config.Profile) *Client {
	httpClient := &http.Client{
		Timeout: 45 * time.Second,
	}

	return &Client{
		BaseURL:      DefaultBaseURL,
		Profile:      prof,
		TokenManager: NewTokenManager(prof, httpClient),
		HTTPClient:   httpClient,
	}
}

// Do executes an authenticated HTTP request to the Apple Ads API.
func (c *Client) Do(ctx context.Context, method, endpoint string, query url.Values, reqBody any, dest any) error {
	token, err := c.TokenManager.GetAccessToken(ctx, false)
	if err != nil {
		return fmt.Errorf("authentication error: %w", err)
	}

	reqURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if reqBody != nil {
		jsonBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if c.Profile.OrgID != "" {
		req.Header.Set("X-AP-Context", "orgId="+c.Profile.OrgID)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// If 401 Unauthorized, try invalidating token cache once and retry
	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.TokenManager.GetAccessToken(ctx, true)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+token)
			if retryResp, retryErr := c.HTTPClient.Do(req); retryErr == nil {
				defer retryResp.Body.Close()
				if retryBytes, rerr := io.ReadAll(retryResp.Body); rerr == nil {
					resp = retryResp
					respBytes = retryBytes
				}
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr ErrorResponse
		if err := json.Unmarshal(respBytes, &apiErr); err == nil && len(apiErr.Errors) > 0 {
			var errMsgs []string
			for _, e := range apiErr.Errors {
				if e.Field != "" {
					errMsgs = append(errMsgs, fmt.Sprintf("[%s] %s: %s", e.MessageCode, e.Field, e.Message))
				} else {
					errMsgs = append(errMsgs, fmt.Sprintf("[%s] %s", e.MessageCode, e.Message))
				}
			}
			return fmt.Errorf("API Error (%d): %s", resp.StatusCode, strings.Join(errMsgs, ", "))
		}
		return fmt.Errorf("API Error (%d): %s", resp.StatusCode, string(respBytes))
	}

	if dest != nil {
		if err := json.Unmarshal(respBytes, dest); err != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}
	}

	return nil
}

// GetACLs retrieves the list of organizations accessible with the current credentials.
func (c *Client) GetACLs(ctx context.Context) ([]ACLRecord, error) {
	var resp ACLResponse
	// The /acls endpoint doesn't require orgId header
	endpoint := "/acls"
	token, err := c.TokenManager.GetAccessToken(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	httpResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch ACLs (status %d): %s", httpResp.StatusCode, string(bodyBytes))
	}

	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
