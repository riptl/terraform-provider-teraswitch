package tsw

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewClient(client *http.Client, baseURL string, apiToken string) *Client {
	return &Client{
		client:  client,
		baseURL: baseURL,
		token:   apiToken,
	}
}

func (c *Client) doForJson(req *http.Request, out any) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var status Status
		if err = json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return resp, fmt.Errorf("unexpected response status (%s)", resp.Status)
		}
		if resp.StatusCode == http.StatusNotFound {
			return resp, ErrNotFound
		}
		return resp, fmt.Errorf("unexpected response status (%d): %s", resp.StatusCode, status.Message)
	}

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return resp, fmt.Errorf("unable to decode response body: %w", err)
	}

	return resp, nil
}
