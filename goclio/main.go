package goclio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	libraryVersion = "0.1"
	userAgent      = "go-clio-client/" + libraryVersion
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

	Account *Account
	Common  *Common
	Matter  *Matter
}

func NewClient() *Client {

	baseUrl := fmt.Sprintf("%s/", APIUrl)

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

func (c *Client) request(method string, path string, data interface{}, v interface{}) error {
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

	resp, err := c.doer.Do(req.WithContext(context.Background()))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	// Return error from facebook API
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
		for k, v := range resp.Header {
			for _, s := range v {
				if k == "Set-Cookie" {
					c.clioCookie += s
				}
			}
		}
	} else if strings.ContainsAny(path, "iris/clio?matter_id=") {
		loc := resp.Request.URL.String()
		jsonStruct := fmt.Sprintf(`{"location":"%s"}`, loc)
		resp.Body = ioutil.NopCloser(bytes.NewBufferString(jsonStruct))
	}

	res := v
	err = json.NewDecoder(resp.Body).Decode(res)

	by, _ := json.Marshal(res)
	if c.EnableLog {
		fmt.Printf("Response %s from %s : %s \n", method, u.String(), string(by))
	}

	return err
}
