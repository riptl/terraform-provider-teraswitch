package tsw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	PowerStateOn  string = "On"
	PowerStateOff string = "off"
)

type Instance struct {
	Id          int64        `json:"id"`
	ObjectType  string       `json:"objectType"`
	PowerState  string       `json:"powerState"`
	IpAddresses []string     `json:"ipAddresses"`
	Tier        InstanceTier `json:"tier"`
	ProjectId   int64        `json:"projectId"`
	ServiceType string       `json:"serviceType"`
	Status      string       `json:"status"`
	RegionId    string       `json:"regionId"`
	TierId      string       `json:"tierId"`
	ImageId     string       `json:"imageId"`
	DisplayName string       `json:"displayName"`
	Region      Region       `json:"region"`
	Sku         string       `json:"sku"`
}

type InstanceTier struct {
	Id       string `json:"id"`
	Memory   int    `json:"memory"`
	Vcpus    int    `json:"vcpus"`
	Transfer int    `json:"transfer"`
}

type Region struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Location string `json:"location"`
}

type InstanceCreateRequest struct {
	DisplayName string   `json:"displayName"`
	RegionId    string   `json:"regionId"`
	TierId      string   `json:"tierId"`
	ImageId     string   `json:"imageId"`
	SshKeyIds   []uint64 `json:"sshKeyIds,omitempty"`
	BootSize    int      `json:"bootSize,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (c *Client) GetInstance(ctx context.Context, id int64) (*Instance, error) {
	uri := c.baseURL + "/v2/Instance/" + strconv.FormatInt(id, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+c.token)

	var result struct {
		Result *Instance `json:"result"`
	}
	_, err = c.doForJson(req, &result)
	if result.Result == nil {
		return nil, fmt.Errorf("unable to get instance")
	}
	return result.Result, err
}

func (c *Client) CreateInstance(ctx context.Context, params *InstanceCreateRequest) (*Instance, error) {
	buf, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	uri := c.baseURL + "/v2/Instance"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+c.token)
	req.Header.Set("content-type", "application/json")

	var result struct {
		Success bool      `json:"success"`
		Result  *Instance `json:"result"`
		Message string    `json:"message"`
	}
	_, err = c.doForJson(req, &result)
	if err != nil {
		return nil, err
	}
	if result.Result == nil || !result.Success {
		return nil, fmt.Errorf("unable to create instance (%s)", result.Message)
	}
	return result.Result, err
}
