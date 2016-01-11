package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://whisk.com" // TODO :: insert real url
)

type Client struct {
	client  *http.Client
	BaseURL *url.URL

	// TODO :: put state in here
	// authToken string // etc.
	// version string
	// verbose bool

	Sdk        *SdkService
	Trigger    *TriggerService
	Action     *ActionService
	Rule       *RuleService
	Activation *ActivationService
	Package    *PackageService
}

func NewClient(httpClient *http.Client) (c *Client, err error) {

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, err := url.Parse(defaultBaseURL)

	c = &Client{
		client:  httpClient,
		BaseURL: baseURL,
	}

	c.Sdk = &SdkService{client: c}
	c.Trigger = &TriggerService{client: c}
	c.Action = &ActionService{client: c}
	c.Rule = &RuleService{client: c}
	c.Activation = &ActivationService{client: c}
	c.Package = &PackageService{client: c}

	return
}

///////////////////////////////
// Request/Utility Functions //
///////////////////////////////

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// Do sends an API request and returns the API response.  The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.  If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}
	return resp, err
}

////////////
// Errors //
////////////

type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
	Errors   []Error        `json:"errors"`  // more detail on individual errors
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v %+v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message, r.Errors)
}

type Error struct {
	Resource string `json:"resource"` // resource on which the error occurred
	Field    string `json:"field"`    // field on which the error occurred
	Code     string `json:"code"`     // validation error code
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v error caused by %v field on %v resource",
		e.Code, e.Field, e.Resource)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}

////////////////////////////
// Basic Client Functions //
////////////////////////////

// Auth performs authorization operation --> stores token in client
func (c *Client) Auth(authKey string) error {
	// Does auth, stores token in client
	return nil
}

// Clean resets object state (cache + auth)
func (c *Client) Clean() {

}

// Version returns the version of the API
func (c *Client) Version() string {
	return ""
}

//List returns lists of all actions, triggers, rules, and activations.
func (c *Client) List() (actions []Action, triggers []Trigger, rules []Rule, activations []Activation, err error) {
	actions, err = c.Action.List()
	if err != nil {
		return
	}

	triggers, err = c.Trigger.List()
	if err != nil {
		return
	}

	rules, err = c.Rule.List()
	if err != nil {
		return
	}

	activations, err = c.Activation.List()
	if err != nil {
		return
	}

	return
}
