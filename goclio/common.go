package goclio

import (
	"fmt"
	"lawatyourside/go-clio-client/goclio/datamodels"
	"net/url"
)

type Common struct {
	client *Client
}

type SearchResponse struct {
	Data *datamodels.SearchResult `json:"data"`
}

func (c *Common) Search(query string) (SearchResponse, error) {
	res := new(SearchResponse)

	// search only matter for now
	qs, _ := url.ParseQuery("")
	qs.Add("fields", "matters{id,highlights,client,status,display_number,description}")
	qs.Add("highlight_end", "{{end}}")
	qs.Add("highlight_start", "{{start}}")
	qs.Add("query", query)
	qs.Add("limit", "10")

	endpoint := fmt.Sprintf("api/v4/search.json?%s", qs.Encode())

	err := c.client.requestJSON("GET", endpoint, nil, res)
	return *res, err
}
