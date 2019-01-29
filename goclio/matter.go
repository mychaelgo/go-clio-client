package goclio

import (
	"fmt"
	"lawatyourside/go-clio-client/goclio/datamodels"
)

type Matter struct {
	client *Client
}

// TODO
func (c *Matter) GetMatter() {

}

type DocumentResponse struct {
	TotalItems int                   `json:"total_items"`
	Children   []datamodels.Document `json:"children"`
	Parents    []datamodels.Document `json:"parents"`
}

func (c *Matter) GetDocuments(folderId string) (DocumentResponse, error) {
	res := new(DocumentResponse)
	endpoint := fmt.Sprintf("iris/folders/" + folderId + "/list.json?limit=25")

	err := c.client.request("POST", endpoint, nil, res)
	return *res, err
}
