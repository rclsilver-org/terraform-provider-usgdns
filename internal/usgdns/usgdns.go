// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package usgdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	usgdns "github.com/rclsilver-org/usg-dns-api/db"
)

type Client struct {
	url   string
	token string
}

func NewClient(url, token string) (*Client, error) {
	return &Client{
		url:   strings.TrimSuffix(url, "/"),
		token: token,
	}, nil
}

func (c *Client) do(method, uri string, body any) (*http.Response, error) {
	parsedURL, err := url.Parse(c.url + uri)
	if err != nil {
		return nil, fmt.Errorf("unable to parse the URL: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal the body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("unable to build the request: %w", err)
	}
	req.Header.Set("Authorization", c.token)

	return http.DefaultClient.Do(req)
}

func (c *Client) GetRecords() ([]usgdns.Record, error) {
	res, err := c.do(http.MethodGet, "/records", nil)
	if err == nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	if err != nil {
		return nil, fmt.Errorf("error while executing the request: %w", err)
	}

	var records []usgdns.Record
	if err := unmarshal(res, &records); err != nil {
		return nil, fmt.Errorf("unable to get the result: %w", err)
	}

	return records, nil
}

func (c *Client) CreateRecord(name, target string) (usgdns.Record, error) {
	res, err := c.do(http.MethodPost, "/records", usgdns.Record{
		Name:   name,
		Target: target,
	})
	if err == nil && res.StatusCode != http.StatusCreated {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)

		errMsg, err2 := getError(res)
		if err2 == nil && errMsg != "" {
			err = fmt.Errorf("%w: %s", err, errMsg)
		}
	}
	if err != nil {
		return usgdns.Record{}, fmt.Errorf("error while executing the request: %w", err)
	}

	var record usgdns.Record
	if err := unmarshal(res, &record); err != nil {
		return usgdns.Record{}, fmt.Errorf("unable to get the result: %w", err)
	}

	return record, nil
}

func (c *Client) GetRecord(id string) (usgdns.Record, error) {
	res, err := c.do(http.MethodGet, "/records/"+id, nil)
	if err == nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)

		errMsg, err2 := getError(res)
		if err2 == nil && errMsg != "" {
			err = fmt.Errorf("%w: %s", err, errMsg)
		}
	}
	if err != nil {
		return usgdns.Record{}, fmt.Errorf("error while executing the request: %w", err)
	}

	var record usgdns.Record
	if err := unmarshal(res, &record); err != nil {
		return usgdns.Record{}, fmt.Errorf("unable to get the result: %w", err)
	}

	return record, nil
}

func (c *Client) UpdateRecord(id, name, target string) (usgdns.Record, error) {
	res, err := c.do(http.MethodPut, "/records/"+id, usgdns.Record{
		Name:   name,
		Target: target,
	})
	if err == nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)

		errMsg, err2 := getError(res)
		if err2 == nil && errMsg != "" {
			err = fmt.Errorf("%w: %s", err, errMsg)
		}
	}
	if err != nil {
		return usgdns.Record{}, fmt.Errorf("error while executing the request: %w", err)
	}

	var record usgdns.Record
	if err := unmarshal(res, &record); err != nil {
		return usgdns.Record{}, fmt.Errorf("unable to get the result: %w", err)
	}

	return record, nil
}

func (c *Client) DeleteRecord(id string) error {
	res, err := c.do(http.MethodDelete, "/records/"+id, nil)
	if err == nil && res.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("error while executing the request: %w", err)
	}

	return nil
}

func unmarshal(res *http.Response, ret any) error {
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read the body: %w", err)
	}
	if err := json.Unmarshal(bodyBytes, &ret); err != nil {
		return fmt.Errorf("unable to unmarshal the body: %w", err)
	}
	return nil
}

func getError(res *http.Response) (string, error) {
	var ret struct {
		Message string `json:"message"`
	}
	if err := unmarshal(res, &ret); err != nil {
		return "", err
	}
	return ret.Message, nil
}
