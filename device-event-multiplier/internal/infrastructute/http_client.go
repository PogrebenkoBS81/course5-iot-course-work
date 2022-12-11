package infrastructure

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// statusCodes - OK status codes
var statusCodes = map[int]bool{
	http.StatusAccepted: true,
	http.StatusCreated:  true,
	http.StatusOK:       true,
}

// NetHTTP - simple http client
type NetHTTP struct {
	client *http.Client
}

// NewHTTPClient - returns new simple http client
func NewHTTPClient(insecureSkipVerify bool, millisecondTimeout int) *NetHTTP {
	return &NetHTTP{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
				},
			},
			Timeout: time.Millisecond * time.Duration(millisecondTimeout),
		},
	}
}

// SendRequest - makes request to given handler with given parameters
func (n *NetHTTP) SendRequest(method, path string, headers http.Header, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header = headers

	res, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { n.closeResponse(res) }()

	if code := res.StatusCode; !statusCodes[code] {
		bts, _ := ioutil.ReadAll(res.Body)

		return nil, fmt.Errorf("error status code: %d, resp: %s", code, string(bts))
	}

	return ioutil.ReadAll(res.Body)
}

// setHeaders - sets request headers
func (n *NetHTTP) setHeaders(req *http.Request, headers map[string]string) {
	for header, value := range headers {
		req.Header.Set(header, value)
	}
}

// closeResponse - response close helpers.
func (n *NetHTTP) closeResponse(res *http.Response) {
	_, _ = io.Copy(ioutil.Discard, res.Body)
	_ = res.Body.Close()
}
