package goopla

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultBaseURL = "https://api.zoopla.co.uk/api/v1/"
)

type Client struct {
	client      *http.Client
	BaseURL     *url.URL
	Credentials Credentials

	Listing *ListingService
	Stream  *StreamService
}

type Credentials struct {
	ApiKey string
}

func newClient() *Client {
	createLogging()

	baseURL, _ := url.Parse(defaultBaseURL)
	client := &Client{client: &http.Client{}, BaseURL: baseURL}

	client.Listing = &ListingService{client: client}
	client.Stream = &StreamService{client: client}

	return client
}

func createLogging() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func NewClient(credentials Credentials, opts ...Opt) (*Client, error) {
	client := newClient()
	client.Credentials = credentials

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) NewRequest(path string, form url.Values) (*http.Request, error) {
	u, err := c.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := DoRequestWithClient(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return resp, err
			}
		} else {
			err = xml.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return resp, err
			}
		}
	}

	return resp, nil
}

func DoRequestWithClient(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}

func CheckResponse(r *http.Response) error {
	if r.StatusCode != 200 {
		err := fmt.Errorf("wrong status code: %d", r.StatusCode)
		return err
	} else {
		return nil
	}
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		if len(v[0]) != 0 || v[0] == "false" {
			origValues[strings.ToLower(k)] = v
		}
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}

type ListingOptions struct {
	postCode       string
	Area           string
	Order_by       string
	Ordering       string
	Listing_status string
	Include_sold   string
	Include_rented string
	Minimum_price  string
	Maximum_price  string
	Minimum_beds   int
	Maximum_beds   int
	Furnished      string
	Property_type  string
	New_homes      bool
	Chain_free     bool
	Keywords       []string
	Listing_id     string
	Branch_id      string
	Page_number    int
	Page_size      int
	Summarized     string
}
