package appleads

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListCampaigns retrieves campaigns with pagination.
func (c *Client) ListCampaigns(ctx context.Context, limit, offset int) ([]Campaign, *Pagination, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}

	var resp CampaignListResponse
	err := c.Do(ctx, http.MethodGet, "/campaigns", query, nil, &resp)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

// GetCampaign fetches a single campaign by ID.
func (c *Client) GetCampaign(ctx context.Context, id int64) (*Campaign, error) {
	endpoint := fmt.Sprintf("/campaigns/%d", id)
	var resp CampaignResponse
	err := c.Do(ctx, http.MethodGet, endpoint, nil, nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
