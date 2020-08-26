package keto

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

type Client struct {
	KetoURL        url.URL
	HTTPClient     *http.Client
	ForwardedProto string
}

func (c *Client) newRequest(method, relativePath string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	u := c.KetoURL
	u.Path = path.Join(u.Path, relativePath)

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if c.ForwardedProto != "" {
		req.Header.Add("X-Forwarded-Proto", c.ForwardedProto)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if os.Getenv("ORIGIN") != "" {
		req.Header.Set("Origin", os.Getenv("ORIGIN"))
	}

	return req, nil

}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if v != nil && resp.StatusCode < 300 {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}
