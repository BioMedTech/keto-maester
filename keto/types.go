package keto

import "encoding/json"

type Flavour string

const (
	Exact Flavour = "exact"
	Regex         = "regex"
	Glob          = "glob"
)

// PolicyJSON represents an Keto policy digestible by ORY Keto
type PolicyJSON struct {
	Id          string          `json:"id"`
	Actions     []string        `json:"actions,omitempty"`
	Conditions  json.RawMessage `json:"conditions,omitempty"`
	Description string          `json:"description,omitempty"`
	Effect      string          `json:"effect,omitempty"`
	Resources   []string        `json:"resources,omitempty"`
	Subjects    []string        `json:"subjects,omitempty"`
}

type Role struct {
	Id      string   `json:"id"`
	Members []string `json:"members,omitempty"`
}

type AddRoleMember struct {
	Members []string `json:"members,omitempty"`
}
