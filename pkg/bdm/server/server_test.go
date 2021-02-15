package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestReplacePort(t *testing.T) {
	res1 := replacePort("myhost.com:1234", 2323)
	if res1 != "myhost.com:2323" {
		t.Fatal(res1)
	}

	res2 := replacePort("127.0.0.1:80", 65000)
	if res2 != "127.0.0.1:65000" {
		t.Fatal(res2)
	}

	res3 := replacePort("subsubdomian.subdomain.server.org", 4242)
	if res3 != "subsubdomian.subdomain.server.org:4242" {
		t.Fatal(res3)
	}

	res4 := replacePort("127.0.0.1", 1)
	if res4 != "127.0.0.1:1" {
		t.Fatal(res4)
	}
}

type FakeResponse struct {
	Status  int
	Headers http.Header
	Data    string
}

func (r *FakeResponse) Write(data []byte) (int, error) {
	r.Data = r.Data + string(data)
	return len(data), nil
}

func (r *FakeResponse) WriteHeader(statusCode int) {
	r.Status = statusCode
}

func (r *FakeResponse) Header() http.Header {
	return r.Headers
}

func TestCreateFallbackHandler(t *testing.T) {
	handler := createFallbackHandler(2323)

	url, _ := url.Parse("/path")
	response1 := FakeResponse{
		Headers: make(http.Header),
	}
	request1 := http.Request{
		Host:   "myhostname.com",
		Method: "GET",
		URL:    url,
	}

	handler(&response1, &request1)
	if response1.Headers["Location"][0] != "https://myhostname.com:2323/path" {
		t.Fatal(response1)
	}

	response2 := FakeResponse{
		Headers: make(http.Header),
	}
	request2 := http.Request{
		Host:   "myhostname.com",
		Method: "POST",
		URL:    url,
	}

	handler(&response2, &request2)
	if response2.Status != 400 || !strings.Contains(response2.Data, "Use HTTPS") {
		t.Fatal(response2)
	}
}
