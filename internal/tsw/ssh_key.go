package tsw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type SshKey struct {
	Id          int64  `json:"id"`
	ProjectId   int64  `json:"projectId"`
	DisplayName string `json:"displayName"`
	SshKey      string `json:"key"`
}

type SshKeyCreateRequest struct {
	DisplayName string `json:"displayName"`
	SshKey      string `json:"key"`
}

func (c *Client) GetSshKey(ctx context.Context, id int64) (*SshKey, error) {
	uri := c.baseURL + "/v1/SSHKey/" + strconv.FormatInt(id, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+c.token)

	var result struct {
		Result *SshKey `json:"result"`
	}
	_, err = c.doForJson(req, &result)
	if result.Result == nil {
		return nil, fmt.Errorf("unable to get ssh key")
	}
	return result.Result, err
}

func (c *Client) CreateSshKey(ctx context.Context, params *SshKeyCreateRequest) (*SshKey, error) {
	buf, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	uri := c.baseURL + "/v1/SSHKey"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+c.token)
	req.Header.Set("content-type", "application/json")

	var result struct {
		Status
		Result *SshKey `json:"result"`
	}
	if _, err = c.doForJson(req, &result); err != nil {
		return nil, err
	}
	if !result.Success || result.Result == nil {
		return nil, fmt.Errorf("unable to create ssh key: message=%s", result.Message)
	}
	return result.Result, err
}

// TODO: TeraSwitch API doesn't properly support DELETE
func (c *Client) DeleteSshKey(ctx context.Context, id int64) error {
	uri := fmt.Sprintf("%s/v1/SSHKey/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Bearer "+c.token)

	key := new(Status)
	if _, err = c.doForJson(req, key); err != nil {
		return err
	}

	if !key.Success {
		return fmt.Errorf("unable to delete ssh key: message=%s", key.Message)
	}
	return nil
}
