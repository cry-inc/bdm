package server

import (
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

func TestCreateFallbackHandler(t *testing.T) {
	handler := createFallbackHandler(2323)

	request1 := createMockedRequest("GET", "/path", nil, nil)
	request1.Host = "myhostname.com"
	response1 := createMockedResponse()
	handler(response1, request1)
	if response1.headers["Location"][0] != "https://myhostname.com:2323/path" {
		t.Fatal(response1)
	}

	request2 := createMockedRequest("POST", "/path", nil, nil)
	request2.Host = "myhostname.com"
	response2 := createMockedResponse()
	handler(response2, request2)
	if response2.status != 400 || !strings.Contains(string(response2.data), "Use HTTPS") {
		t.Fatal(response2)
	}
}
