package goclio

import (
	"fmt"
	"lawatyourside/go-clio-client/goclio/datamodels"
	"net/url"
)

type Matter struct {
	client *Client
}

// TODO
func (c *Matter) GetMatter() {

}

type GetLocationResponse struct {
	Location string `json:"location"`
}

// returned folderId
func (c *Matter) GetLocationDocuments(matterId int) (string, error) {
	res := new(GetLocationResponse)

	folderId := ""

	endpoint := fmt.Sprintf("iris/clio?matter_id=%d", matterId)
	err := c.client.request("GET", endpoint, nil, res)
	if err != nil {
		return folderId, err
	}

	u, err := url.Parse(res.Location)
	if err != nil {
		return folderId, err
	}

	m, _ := url.ParseQuery(u.Path + u.Fragment)

	folderId = m.Get("/iris//drive/?id")

	return folderId, err
}

type DocumentsResponse struct {
	TotalItems int                   `json:"total_items"`
	Children   []datamodels.Document `json:"children"`
	Parents    []datamodels.Document `json:"parents"`
}

func (c *Matter) GetDocuments(folderId string) (DocumentsResponse, error) {
	res := new(DocumentsResponse)
	endpoint := fmt.Sprintf("iris/folders/" + folderId + "/list.json?limit=25")

	err := c.client.request("GET", endpoint, nil, res)
	return *res, err
}

type DocumentResponse struct {
	Item datamodels.Document `json:"item"`
}

func (c *Matter) GetDocument(documentId string) (DocumentResponse, error) {
	res := new(DocumentResponse)
	endpoint := fmt.Sprintf("iris/items/" + documentId + ".json")

	err := c.client.request("GET", endpoint, nil, res)
	return *res, err
}
