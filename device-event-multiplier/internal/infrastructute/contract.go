package infrastructure

import "net/http"

type HTTPClient interface {
	SendRequest(method, path string, headers http.Header, body []byte) ([]byte, error)
}
