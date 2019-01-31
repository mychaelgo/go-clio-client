package goclio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"lawatyourside/go-clio-client/goclio/utils"
	"net/http"
	"net/url"
	"strings"
)

const (
	libraryVersion = "0.1"
	//userAgent      = "go-clio-client/" + libraryVersion
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"
)

var (
	APIUrl = "https://app.clio.com/"
)

var (
	// ErrUnauthorized can be returned on any call on response status code 401.
	ErrUnauthorized = errors.New("go-clio-client: unauthorized")
)

type errorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Message string `json:"message,omitempty"`
}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DoerFunc func(req *http.Request) (resp *http.Response, err error)

type Client struct {
	doer       doer
	baseURL    *url.URL
	userAgent  string
	httpClient *http.Client
	EnableLog  bool

	clioCookie string
	csrfToken  string
	xsrfToken  string

	Account *Account
	Common  *Common
	Matter  *Matter
}

func NewClient() *Client {

	baseUrl := fmt.Sprintf("%s", APIUrl)

	baseURL, _ := url.Parse(baseUrl)
	client := &Client{
		doer:      http.DefaultClient,
		baseURL:   baseURL,
		userAgent: userAgent,
	}

	client.Account = &Account{client}
	client.Common = &Common{client}
	client.Matter = &Matter{client}

	return client
}

func (c *Client) SetCookie(s string) {
	c.clioCookie = s
}

func (c *Client) GetCookie() string {
	return c.clioCookie
}

func (c *Client) requestJSON(method string, path string, data interface{}, v interface{}) error {
	urlStr := path

	rel, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	u := c.baseURL.ResolveReference(rel)
	var body io.Reader

	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)

	}

	if c.EnableLog {
		fmt.Printf("Request %s to %s with data: %s \n", method, u.String(), body)
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cookie", c.clioCookie)
	req.Header.Set("X-CSRF-Token", c.csrfToken)
	req.Header.Set("X-XSRF-TOKEN", c.xsrfToken)

	resp, err := c.doer.Do(req.WithContext(context.Background()))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	responseString := string(responseData)

	if c.EnableLog {
		fmt.Printf("Response %s from %s : %s \n", method, u.String(), responseString)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	// Return error from clio API
	if resp.StatusCode != http.StatusOK {
		rb := new(errorResponse)

		err = json.NewDecoder(resp.Body).Decode(rb)

		if err != nil {
			return errors.New("general error")
		}

		return errors.New(rb.Error.Message)
	}

	// Decode to interface
	if path == "session.json" && resp.Body != nil {
		var cookieMaps []utils.CookieMap
		cookieMaps = append(cookieMaps, utils.CookieStringToMap(c.clioCookie))
		for k, v := range resp.Header {
			for _, s := range v {
				if k == "Set-Cookie" {
					cookieMap := utils.CookieStringToMap(s)
					cookieMaps = append(cookieMaps, cookieMap)
				}
			}
		}
		//cookieMaps = append(cookieMaps, utils.CookieStringToMap(c.clioCookie))
		mergedCookieMap := utils.MergeCookieMap(cookieMaps...)
		if mergedCookieMap["XSRF-TOKEN"] != "" {
			c.xsrfToken = mergedCookieMap["XSRF-TOKEN"]
		}
		c.clioCookie = utils.CookieMapToString(mergedCookieMap)
		return nil
	} else if strings.Contains(path, "iris/clio?matter_id=") {
		loc := resp.Request.URL.String()
		jsonStruct := fmt.Sprintf(`{"location":"%s"}`, loc)
		resp.Body = ioutil.NopCloser(bytes.NewBufferString(jsonStruct))
		return nil
	}

	//err = json.NewDecoder(resp.Body).Decode(res)
	_ = json.Unmarshal([]byte(responseString), &v)

	//res := v
	//err = json.NewDecoder(resp.Body).Decode(res)
	//
	//by, _ := json.Marshal(res)
	//if c.EnableLog {
	//	fmt.Printf("Response %s from %s : %s \n", method, u.String(), string(by))
	//}

	return nil
}

func (c *Client) requestFormEncoded(method string, path string, payload url.Values) error {
	urlStr := path

	rel, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	u := c.baseURL.ResolveReference(rel)

	if c.EnableLog {
		fmt.Printf("Request %s to %s with data: %s \n", method, u.String(), payload)
	}

	req, err := http.NewRequest(method, u.String(), strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}

	//fmt.Println("payload ", payload.Encode())

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	//req.Header.Set("Host", "app.clio.com")
	req.Header.Set("Origin", "https://app.clio.com")
	req.Header.Set("Referer", "https://app.clio.com/matters")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cookie", c.clioCookie)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-CSRF-Token", c.csrfToken)
	req.Header.Set("X-XSRF-TOKEN", c.xsrfToken)
	//req.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))

	resp, err := c.doer.Do(req.WithContext(context.Background()))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	}
	responseString := string(responseData)

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	// Return error from clio API
	if resp.StatusCode != http.StatusOK {
		fmt.Println("responseString ", responseString)
		return errors.New(fmt.Sprintf("error %d , ", resp.StatusCode))
	}

	// set cookie XSRF-TOKEN
	var cookieMaps []utils.CookieMap
	cookieMaps = append(cookieMaps, utils.CookieStringToMap(c.clioCookie))
	for k, v := range resp.Header {
		for _, s := range v {
			if k == "Set-Cookie" {
				cookieMap := utils.CookieStringToMap(s)
				cookieMaps = append(cookieMaps, cookieMap)
			}
		}
	}
	//cookieMaps = append(cookieMaps, utils.CookieStringToMap(c.clioCookie))
	mergedCookieMap := utils.MergeCookieMap(cookieMaps...)
	if mergedCookieMap["XSRF-TOKEN"] != "" {
		c.xsrfToken = mergedCookieMap["XSRF-TOKEN"]
	}
	c.clioCookie = utils.CookieMapToString(mergedCookieMap)

	//fmt.Println("requestFormEncoded Cookie ", c.clioCookie)
	//fmt.Println("requestFormEncoded csrfToken ", c.csrfToken)

	return err
}

func (c *Client) Scrapper(path string) (*io.ReadCloser, error) {
	return c.scrapper(path)
}

func (c *Client) scrapper(path string) (*io.ReadCloser, error) {
	urlStr := path

	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	u := c.baseURL.ResolveReference(rel)

	if c.EnableLog {
		fmt.Printf("Scrape url %s \n", u.String())
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	//fmt.Println("scrapper Cookie ", c.clioCookie)

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Cookie", c.clioCookie)
	req.Header.Set("X-CSRF-Token", c.csrfToken)
	req.Header.Set("X-XSRF-TOKEN", c.xsrfToken)

	resp, err := c.doer.Do(req.WithContext(context.Background()))
	if err != nil {
		return nil, err
	}

	//defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	//fmt.Println("scrapper Cookie ", c.clioCookie)
	//fmt.Println("scrapper csrfToken ", c.csrfToken)

	// set cookie XSRF-TOKEN
	var cookieMaps []utils.CookieMap
	cookieMaps = append(cookieMaps, utils.CookieStringToMap(c.clioCookie))
	for k, v := range resp.Header {
		for _, s := range v {
			if k == "Set-Cookie" {
				cookieMap := utils.CookieStringToMap(s)
				cookieMaps = append(cookieMaps, cookieMap)
			}
		}
	}
	//cookieMaps = append(cookieMaps, utils.CookieStringToMap(c.clioCookie))
	mergedCookieMap := utils.MergeCookieMap(cookieMaps...)
	if mergedCookieMap["XSRF-TOKEN"] != "" {
		c.xsrfToken = mergedCookieMap["XSRF-TOKEN"]
	}
	c.clioCookie = utils.CookieMapToString(mergedCookieMap)

	//fmt.Println("scrapper Cookie ", c.clioCookie)

	return &resp.Body, err
}
