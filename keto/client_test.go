package keto_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ory/keto-maester/keto"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	policiesEndpoint = "/clients"
	schemeHTTP       = "http"

	testID                          = "test-id"
	testClient                      = `{"id":"test-id"}`
	testClientCreated               = `{"id":"test-id-1","actions":["create","update"],"description":"Grant permission for user maria","effect":"allow","resources":["resources:articles"],"subjects":["users:maria"]}`
	testClientWithConditionsCreated = `{"id":"test-id-1","actions":["create","update"],"description":"Grant permission for user maria","effect":"allow","resources":["resources:articles"],"subjects":["users:maria"]},"conditions":{"foo":{"type":"StringMatchCondition","options":{"StringMatchCondition":"bar.+"}}}`

	statusNotFoundBody            = `{"error":"Not Found","error_description":"Unable to locate the requested resource","status_code":404,"request_id":"id"}`
	statusInternalServerErrorBody = "the server encountered an internal error or misconfiguration and was unable to complete your request"
)

type server struct {
	statusCode int
	respBody   string
	err        error
}

var testPolicyUpsert = &keto.PolicyJSON{
	Actions:     []string{"create", "update"},
	Description: "Grant permission for user maria",
	Effect:      "allow",
	Resources:   []string{"resources:articles"},
	Subjects:    []string{"users:maria"},
}

func TestCRUD(t *testing.T) {

	assert := assert.New(t)

	c := keto.Client{
		HTTPClient: &http.Client{},
		KetoURL:    url.URL{Scheme: schemeHTTP},
	}

	t.Run("method=get", func(t *testing.T) {

		for d, tc := range map[string]server{
			"getting registered policy": {
				http.StatusOK,
				testClient,
				nil,
			},
			"getting unregistered policy": {
				http.StatusNotFound,
				statusNotFoundBody,
				nil,
			},
			"internal server error when requesting": {
				http.StatusInternalServerError,
				statusInternalServerErrorBody,
				errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {

				//given
				shouldFind := tc.statusCode == http.StatusOK

				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(fmt.Sprintf("%s%s", c.KetoURL.String(), c.AcpEnginePolicyPath(keto.Exact, testID)), fmt.Sprintf("%s://%s%s", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodGet, req.Method)
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.respBody))
					if shouldFind {
						w.Header().Set("Content-type", "application/json")
					}
				})
				runServer(&c, h)

				//when
				o, found, err := c.GetPolicy(keto.Exact, testID)

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}

				assert.Equal(shouldFind, found)
				if shouldFind {
					require.NotNil(t, o)
					var expected keto.PolicyJSON
					json.Unmarshal([]byte(testClient), &expected)
					assert.Equal(&expected, o)
				}
			})
		}
	})

	t.Run("method=put", func(t *testing.T) {

		for d, tc := range map[string]server{
			"with new policy": {
				http.StatusOK,
				testClientCreated,
				nil,
			},
			"with new policy with condition": {
				http.StatusOK,
				testClientWithConditionsCreated,
				nil,
			},
			"with updating existing policy": {
				http.StatusOK,
				testClientCreated,
				nil,
			},
			"internal server error when requesting": {
				http.StatusInternalServerError,
				statusInternalServerErrorBody,
				errors.New("http request returned unexpected status code"),
			},
		} {
			t.Run(fmt.Sprintf("case/%s", d), func(t *testing.T) {
				var (
					err      error
					o        *keto.PolicyJSON
					expected *keto.PolicyJSON
				)
				//given
				new := tc.statusCode == http.StatusOK
				newWithConditions := d == "with new policy with conditions"

				h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(fmt.Sprintf("%s%s", c.KetoURL.String(), c.AcpEnginePolicyPath(keto.Exact, "")), fmt.Sprintf("%s://%s%s/", schemeHTTP, req.Host, req.URL.Path))
					assert.Equal(http.MethodPost, req.Method)
					w.WriteHeader(tc.statusCode)
					w.Write([]byte(tc.respBody))
					if new {
						w.Header().Set("Content-type", "application/json")
					}
				})
				runServer(&c, h)

				//when
				if newWithConditions {
					conditions, _ := json.Marshal(map[string]interface{}{
						"foo": map[string]interface{}{
							"type": "StringMatchCondition",
							"options": map[string]string{
								"StringMatchCondition": "bar.+",
							},
						},
					})

					var testPolicyJSONUpsert = &keto.PolicyJSON{
						Actions:     []string{"create", "update"},
						Description: "Grant permission for user maria",
						Effect:      "allow",
						Resources:   []string{"resources:articles"},
						Subjects:    []string{"users:maria"},
						Conditions:  conditions,
					}
					o, err = c.UpsertPolicy(keto.Exact, testPolicyJSONUpsert)
					expected = testPolicyJSONUpsert
				} else {
					o, err = c.UpsertPolicy(keto.Exact, testPolicyUpsert)
					expected = testPolicyUpsert
				}

				//then
				if tc.err == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Contains(err.Error(), tc.err.Error())
				}

				if new {
					require.NotNil(t, o)
					assert.Equal(expected.Actions, o.Actions)
					assert.Equal(expected.Conditions, o.Conditions)
					assert.Equal(expected.Description, o.Description)
					assert.Equal(expected.Effect, o.Effect)
					assert.NotNil(o.Id)
					if newWithConditions {
						assert.NotNil(o.Conditions)
						assert.True(len(o.Conditions) > 0)
						for key, _ := range o.Conditions {
							assert.Equal(o.Conditions[key], expected.Conditions[key])
						}
					} else {
						assert.Nil(o.Conditions)
					}
				}
			})
		}
	})

}

func runServer(c *keto.Client, h http.HandlerFunc) {
	s := httptest.NewServer(h)
	serverUrl, _ := url.Parse(s.URL)
	c.KetoURL = *serverUrl.ResolveReference(&url.URL{Path: policiesEndpoint})
}
