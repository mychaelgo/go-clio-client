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

func (c *Matter) GetMatter(matterId string) (datamodels.Matter, error) {
	res := new(datamodels.Matter)

	// get matter
	fields := "id,display_number,custom_number,description,status,location,client_reference,billable,maildrop_address,billing_method,open_date,close_date,pending_date,client{id,name,first_name,last_name,type,initials},practice_area{id,name},shared,contingency_fee{show_contingency_award},responsible_attorney{id,name,first_name,last_name,enabled},originating_attorney{id,name,first_name,last_name,enabled},group{id,name,type},statute_of_limitations{id,due_at,status,reminders},relationships{id,description,contact}"

	endpoint := fmt.Sprintf("api/v4/matters/%s?fields=%s", matterId, fields)

	err := c.client.requestJSON("GET", endpoint, nil, res)
	return *res, err
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

func (c *Matter) GetDocuments(folderId int) (DocumentsResponse, error) {
	var err error
	res := new(DocumentsResponse)

	endpoint := fmt.Sprintf("iris/folders/%d/list.json?limit=25", folderId)

	err = c.client.requestJSON("GET", endpoint, nil, res)

	return *res, err
}

type DocumentResponse struct {
	Item datamodels.Document `json:"item"`
}

func (c *Matter) GetDocument(documentId int) (DocumentResponse, error) {
	res := new(DocumentResponse)
	endpoint := fmt.Sprintf("iris/items/%d.json", documentId)

	err := c.client.requestJSON("GET", endpoint, nil, res)
	return *res, err
}

type DocumentTemplateResponse struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
}

type DocumentTemplateResponseMap map[string]DocumentTemplateResponse

func (c *Matter) GetDocumentTemplate(matterId int, clientId int) ([]DocumentTemplateResponse, DocumentTemplateResponseMap, error) {
	var res []DocumentTemplateResponse
	endpoint := fmt.Sprintf("export_matter_ddps/new?matter_id=%d&client_id=%d&dt_table_id=", matterId, clientId)

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

func (c *Matter) CreateNewTokenForDocTemplate(matterId int, clientId int) (bool, error) {
	endpoint := fmt.Sprintf("export_matter_ddps/new?matter_id=%d&client_id=%d&dt_table_id=", matterId, clientId)

	_, err := c.client.scrapper(endpoint)
	if err != nil {
		return false, err
	}

	return true, err
}

func (c *Matter) GetDocumentCsrfToken(folderId int) (string, error) {
	var res string
	endpoint := fmt.Sprintf("iris/#/drive?id=%d&parent_offset=2", folderId)

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

func (c *Matter) CreateDocumentTemplate(templateId int, clientId int, matterId int, folderId int, documentType constant.ClioDocType, fileName string) (bool, error) {
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
	payload.Set("export[document_template_id]", string(templateId))
	payload.Set("export[client_id]", string(clientId))
	payload.Set("export_matter_id_auto_complete_input", "00856-Soria")
	payload.Set("export[matter_id]", string(matterId))
	payload.Set("export[matter_id]", string(matterId))
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
