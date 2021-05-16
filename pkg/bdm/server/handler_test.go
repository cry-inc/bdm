package server

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type mockResponseWriter struct {
	status  int
	headers http.Header
	data    []byte
}

func (r *mockResponseWriter) Write(data []byte) (int, error) {
	r.data = append(r.data, data...)
	return len(data), nil
}

func (r *mockResponseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *mockResponseWriter) Header() http.Header {
	return r.headers
}

func createMockedResponse() *mockResponseWriter {
	return &mockResponseWriter{
		headers: make(http.Header),
	}
}

func createMockedRequest(method, path string, body *string, authUser *string) *http.Request {
	url, _ := url.Parse(path)
	request := http.Request{
		Method: method,
		URL:    url,
		Header: make(http.Header),
	}
	if body != nil {
		request.Body = ioutil.NopCloser(strings.NewReader(*body))
	}
	if authUser != nil {
		authToken := createAuthToken(*authUser, defaultExpiration)
		request.Header.Add("Cookie", "login="+authToken.Token)
	}
	return &request
}
