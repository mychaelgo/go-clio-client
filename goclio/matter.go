package goclio

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"lawatyourside/go-clio-client/goclio/constant"
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
	err := c.client.requestJSON("GET", endpoint, nil, res)
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

	err := c.client.requestJSON("GET", endpoint, nil, res)
	return *res, err
}

type DocumentResponse struct {
	Item datamodels.Document `json:"item"`
}

func (c *Matter) GetDocument(documentId string) (DocumentResponse, error) {
	res := new(DocumentResponse)
	endpoint := fmt.Sprintf("iris/items/" + documentId + ".json")

	err := c.client.requestJSON("GET", endpoint, nil, res)
	return *res, err
}

type DocumentTemplateResponse struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

type DocumentTemplateResponseMap map[string]DocumentTemplateResponse

func (c *Matter) GetDocumentTemplate(matterId string, clientId string) ([]DocumentTemplateResponse, DocumentTemplateResponseMap, error) {
	var res []DocumentTemplateResponse
	endpoint := fmt.Sprintf("export_matter_ddps/new?matter_id=" + matterId + "&client_id=" + clientId + "&dt_table_id=")

	body, err := c.client.scrapper(endpoint)
	defer (*body).Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(*body)
	if err != nil {
		return nil, nil, err
	}

	// Find doc template
	mapResponse := DocumentTemplateResponseMap{}
	doc.Find("select option").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the value
		data := DocumentTemplateResponse{
			Name:      s.First().AttrOr("data-basename", ""),
			Extension: s.First().AttrOr("data-extname", ""),
			Value:     s.First().AttrOr("value", ""),
		}
		// add to array
		res = append(res, data)

		// assign to map, for fast searching
		mapResponse[data.Name] = data
	})

	return res, mapResponse, err
}

func (c *Matter) CreateNewTokenForDocTemplate(matterId string, clientId string) (bool, error) {
	endpoint := fmt.Sprintf("export_matter_ddps/new?matter_id=" + matterId + "&client_id=" + clientId + "&dt_table_id=")

	_, err := c.client.scrapper(endpoint)
	if err != nil {
		return false, err
	}

	return true, err
}

func (c *Matter) GetDocumentCsrfToken(folderId string) (string, error) {
	var res string
	endpoint := fmt.Sprintf("iris/#/drive?id=" + folderId + "&parent_offset=2")

	body, err := c.client.scrapper(endpoint)
	defer (*body).Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(*body)
	if err != nil {
		return res, err
	}

	// Find doc template
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the value
		metaValue := s.First().AttrOr("name", "")
		if metaValue == "csrf-token" {
			res = s.First().AttrOr("content", "")
		}
	})

	c.client.csrfToken = res

	return res, err
}

func (c *Matter) CreateDocumentTemplate(templateId string, clientId string, matterId string, folderId string, documentType constant.ClioDocType, fileName string) (bool, error) {
	var err error

	// get current page csrf-token
	_, err = c.GetDocumentCsrfToken(folderId)
	if err != nil {
		return false, err
	}

	_, err = c.CreateNewTokenForDocTemplate(matterId, clientId)
	if err != nil {
		return false, err
	}

	endpoint := fmt.Sprintf("export_matter_ddps")

	generateDocx := "0"
	generatePDF := "0"

	switch documentType {
	case constant.ClioDocTypeDocx:
		generateDocx = "1"
		break
	case constant.ClioDocTypePDF:
		generatePDF = "1"
		break
	}

	payload := url.Values{}
	payload.Set("utf8", fmt.Sprintf("%s", "\u2713")) // ✓
	payload.Set("dt_table_id", "")
	payload.Set("export[document_template_id]", templateId)
	payload.Set("export[client_id]", clientId)
	payload.Set("export_matter_id_auto_complete_input", "00856-Soria")
	payload.Set("export[matter_id]", matterId)
	payload.Set("export[matter_id]", matterId)
	payload.Set("export[export_pdf]", generatePDF)
	payload.Set("export[export_original]", generateDocx)
	payload.Set("export[export_basename_original]", fileName)
	payload.Set("commit", "Create")

	err = c.client.requestFormEncoded("POST", endpoint, payload)
	if err != nil {
		return false, err
	}

	return true, nil
}
