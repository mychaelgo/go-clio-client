package goclio

import (
	"fmt"
	"lawatyourside/go-clio-client/goclio/datamodels"
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
	fields := "matters%7Bid%2Chighlights%2Cclient%2Cstatus%2Cdisplay_number%2Cdescription%7D"
	limit := 10

	endpoint := fmt.Sprintf("api/v4/search.json?%s&%s&%d", fields, query, limit)

	err := c.client.request("GET", endpoint, nil, res)
	return *res, err
}
