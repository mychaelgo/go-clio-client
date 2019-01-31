package goclio

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"lawatyourside/go-clio-client/goclio/constant"
	"lawatyourside/go-clio-client/goclio/datamodels"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Matter struct {
	client *Client
}

type MatterResponse struct {
	Data datamodels.Matter `json:"data"`
}

func (c *Matter) GetMatter(matterId int) (MatterResponse, error) {
	res := new(MatterResponse)

	// get matter
	qs, _ := url.ParseQuery("")
	qs.Add("fields", "id,display_number,custom_number,description,status,location,client_reference,billable,maildrop_address,billing_method,open_date,close_date,pending_date,client{id,name,first_name,last_name,type,initials},practice_area{id,name},shared,contingency_fee{show_contingency_award},responsible_attorney{id,name,first_name,last_name,enabled},originating_attorney{id,name,first_name,last_name,enabled},group{id,name,type},statute_of_limitations{id,due_at,status,reminders},relationships{id,description,contact}")
	endpoint := fmt.Sprintf("api/v4/matters/%d?%s", matterId, qs.Encode())

	err := c.client.requestJSON("GET", endpoint, nil, res)
	return *res, err
}

type GetLocationResponse struct {
	Location string `json:"location"`
}

// returned folderId
func (c *Matter) GetLocationDocuments(matterId int) (int, error) {
	res := new(GetLocationResponse)

	folderId := 0

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

	folderIdInt := m.Get("/iris//drive/?id")
	folderId, _ = strconv.Atoi(folderIdInt)

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

func (c *Matter) CreateDocumentTemplate(templateId int, clientId int, matterId int, folderId int, documentType constant.ClioDocType, fileName string) (success bool, documentId int, err error) {
	documentId = 0

	// get current page csrf-token
	_, err = c.GetDocumentCsrfToken(folderId)
	if err != nil {
		return false, documentId, err
	}

	_, err = c.CreateNewTokenForDocTemplate(matterId, clientId)
	if err != nil {
		return false, documentId, err
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
		return false, documentId, err
	}

	// get newly created document id
	documentId, _ = c.GetNewlyCreatedDocumentId(folderId, fileName)

	return true, documentId, nil
}

func (c *Matter) GetNewlyCreatedDocumentId(folderId int, fileName string) (documentId int, err error) {
	docs, err := c.GetDocuments(folderId)
	now := time.Now().UTC()
	for _, v := range docs.Children {
		if strings.Contains(v.Name, fileName) {
			diff := now.Sub(v.CreatedAt.UTC())
			if diff.Minutes() <= constant.ClioMaxWaitMinutesForDocument {
				documentId = v.Id
				break
			}
		}
	}

	if documentId == 0 {
		return c.GetNewlyCreatedDocumentId(folderId, fileName)
	}

	return documentId, nil
}
