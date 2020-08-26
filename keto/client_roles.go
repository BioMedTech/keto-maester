package keto

import (
	"fmt"
	"net/http"
)

func (c *Client) AcpEngineRolePath(flavour Flavour, id string) string {
	return fmt.Sprintf("/engines/acp/ory/%s/roles/%s", flavour, id)
}

func (c *Client) GetRole(flavour Flavour, id string) (*Role, bool, error) {
	var jsonClient *Role

	req, err := c.newRequest(http.MethodGet, c.AcpEngineRolePath(flavour, id), nil)
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

func (c *Client) ListRole(flavour Flavour) ([]*Role, error) {
	var jsonClientList []*Role

	req, err := c.newRequest(http.MethodGet, c.AcpEngineRolePath(flavour, ""), nil)
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

func (c *Client) UpsertRole(flavour Flavour, o *Role) (*Role, error) {
	var jsonClient *Role

	req, err := c.newRequest(http.MethodPut, c.AcpEngineRolePath(flavour, ""), o)
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

func (c *Client) DeleteRole(flavour Flavour, id string) error {
	req, err := c.newRequest(http.MethodDelete, c.AcpEngineRolePath(flavour, id), nil)
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
