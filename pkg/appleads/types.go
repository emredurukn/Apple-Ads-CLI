package appleads

import "time"

// TokenResponse represents the OAuth2 response from Apple ID token endpoint.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// APIError represents an individual error returned by the Apple Ads API.
type APIError struct {
	MessageCode string `json:"messageCode"`
	Message     string `json:"message"`
	Field       string `json:"field,omitempty"`
}

// ErrorResponse represents the top-level error response payload.
type ErrorResponse struct {
	Errors []APIError `json:"errors"`
}

// Pagination metadata returned in list endpoints.
type Pagination struct {
	TotalResults int `json:"totalResults"`
	StartIndex   int `json:"startIndex"`
	ItemsPerPage int `json:"itemsPerPage"`
}

// ACLRecord represents an account/organization access record.
type ACLRecord struct {
	OrgID     int64    `json:"orgId"`
	OrgName   string   `json:"orgName"`
	RoleNames []string `json:"roleNames"`
	ParentOrg bool     `json:"parentOrg"`
}

// ACLResponse represents response from /api/v5/acls endpoint.
type ACLResponse struct {
	Data       []ACLRecord `json:"data"`
	Pagination Pagination  `json:"pagination,omitempty"`
}

// Money represents monetary amounts with currency.
type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// Campaign represents an Apple Search Ads campaign.
type Campaign struct {
	ID                 int64      `json:"id"`
	OrgID              int64      `json:"orgId"`
	Name               string     `json:"name"`
	AdamID             int64      `json:"adamId,omitempty"`
	BudgetAmount       *Money     `json:"budgetAmount,omitempty"`
	DailyBudgetAmount  *Money     `json:"dailyBudgetAmount,omitempty"`
	BillingCurrency    string     `json:"billingCurrency,omitempty"`
	Status             string     `json:"status"`
	ServingStatus      string     `json:"servingStatus"`
	ServingStateReasons []string  `json:"servingStateReasons,omitempty"`
	CountriesOrRegions []string   `json:"countriesOrRegions,omitempty"`
	StartTime          *time.Time `json:"startTime,omitempty"`
	EndTime            *time.Time `json:"endTime,omitempty"`
	ModificationTime   *time.Time `json:"modificationTime,omitempty"`
}

// CampaignListResponse is the API response wrapper for campaign listings.
type CampaignListResponse struct {
	Data       []Campaign `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// CampaignResponse is the API response wrapper for a single campaign.
type CampaignResponse struct {
	Data Campaign `json:"data"`
}

// AdGroup represents an ad group inside a campaign.
type AdGroup struct {
	ID               int64      `json:"id"`
	CampaignID       int64      `json:"campaignId"`
	OrgID            int64      `json:"orgId"`
	Name             string     `json:"name"`
	DefaultBidAmount *Money     `json:"defaultBidAmount,omitempty"`
	CPA上演Amount    *Money     `json:"cpaGoal,omitempty"`
	Status           string     `json:"status"`
	ServingStatus    string     `json:"servingStatus"`
	ModificationTime *time.Time `json:"modificationTime,omitempty"`
}

// AdGroupListResponse is the API response wrapper for ad groups.
type AdGroupListResponse struct {
	Data       []AdGroup  `json:"data"`
	Pagination Pagination `json:"pagination"`
}
