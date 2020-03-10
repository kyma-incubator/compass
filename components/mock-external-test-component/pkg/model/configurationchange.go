package model

import "time"

type ConfigurationChange struct {
	UUID       string      `json:"uuid"`
	User       string      `json:"user"`
	Time       time.Time   `json:"time"`
	Object     Object      `json:"object"`
	Attributes []Attribute `json:"attributes"`
	Tenant     string      `json:"tenant"`
	Success    *bool       `json:"success"`
}

type Attribute struct {
	Name string `json:"name"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type Object struct {
	ID   map[string]string `json:"id"`
	Type string            `json:"type"`
}

type ID struct {
	ExtraData map[string]string `json:"extra_data"`
}

type SuccessResponse struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
