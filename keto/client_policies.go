package keto

import (
	"fmt"
	"net/http"
)

func (c *Client) AcpEnginePolicyPath(flavour Flavour, id string) string {
	return fmt.Sprintf("/engines/acp/ory/%s/policies/%s", flavour, id)
}

func (c *Client) GetPolicy(flavour Flavour, id string) (*PolicyJSON, bool, error) {
	var jsonClient *PolicyJSON

	req, err := c.newRequest(http.MethodGet, c.AcpEnginePolicyPath(flavour, id), nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := c.do(req, &jsonClient)
	if err != nil {
		return nil, false, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return jsonClient, true, nil
	case http.StatusNotFound:
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("%s %s http request returned unexpected status code %s", req.Method, req.URL.String(), resp.Status)
	}
}

func (c *Client) ListPolicy(flavour Flavour) ([]*PolicyJSON, error) {

	var jsonClientList []*PolicyJSON

	req, err := c.newRequest(http.MethodGet, c.AcpEnginePolicyPath(flavour, ""), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req, &jsonClientList)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return jsonClientList, nil
	default:
		return nil, fmt.Errorf("%s %s http request returned unexpected status code %s", req.Method, req.URL.String(), resp.Status)
	}
}

func (c *Client) UpsertPolicy(flavour Flavour, o *PolicyJSON) (*PolicyJSON, error) {
	var jsonClient *PolicyJSON

	req, err := c.newRequest(http.MethodPut, c.AcpEnginePolicyPath(flavour, ""), o)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req, &jsonClient)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return jsonClient, nil
	default:
		return nil, fmt.Errorf("%s %s http request returned unexpected status code: %s", req.Method, req.URL, resp.Status)
	}
}

func (c *Client) DeletePolicy(flavour Flavour, id string) error {

	req, err := c.newRequest(http.MethodDelete, c.AcpEnginePolicyPath(flavour, id), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		fmt.Printf("client with id %s does not exist", id)
		return nil
	default:
		return fmt.Errorf("%s %s http request returned unexpected status code %s", req.Method, req.URL.String(), resp.Status)
	}
}
