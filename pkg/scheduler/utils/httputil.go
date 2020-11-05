package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

// GetTransport get Transport
func GetTransport(isClientCARequired bool) *http.Transport {
	tr, err := initTransport(isClientCARequired)
	if err != nil {
	}
	return tr
}

// initTransport initialize http
func initTransport(isClientCARequired bool) (*http.Transport, error) {
	tr := &http.Transport{DisableKeepAlives: true}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return tr, nil
}

// SendHTTPRequest send http request
func SendHTTPRequest(method string, serverAddr string, headers map[string][]string,
	body io.Reader, isClientCARequired bool) (r *http.Response, e error) {

	client := &http.Client{Transport: GetTransport(isClientCARequired), Timeout: time.Duration(2 * time.Minute)}

	req, err := http.NewRequest(method, serverAddr, body)
	if err != nil {
		return nil, err
	}
	req.Close = true

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
