package server

import (
	"strings"
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestReplacePort(t *testing.T) {
	res1 := replacePort("myhost.com:1234", 2323)
	util.AssertEqualString(t, "myhost.com:2323", res1)

	res2 := replacePort("127.0.0.1:80", 65000)
	util.AssertEqualString(t, "127.0.0.1:65000", res2)

	res3 := replacePort("subsubdomian.subdomain.server.org", 4242)
	util.AssertEqualString(t, "subsubdomian.subdomain.server.org:4242", res3)

	res4 := replacePort("127.0.0.1", 1)
	util.AssertEqualString(t, "127.0.0.1:1", res4)
}

func TestCreateFallbackHandler(t *testing.T) {
	handler := createFallbackHandler(2323)

	request1 := createMockedRequest("GET", "/path", nil, nil)
	request1.Host = "myhostname.com"
	response1 := createMockedResponse()
	handler(response1, request1)
	util.AssertEqualString(t, "https://myhostname.com:2323/path", response1.headers["Location"][0])

	request2 := createMockedRequest("POST", "/path", nil, nil)
	request2.Host = "myhostname.com"
	response2 := createMockedResponse()
	handler(response2, request2)
	util.Assert(t, response2.status == 400)
	util.Assert(t, strings.Contains(string(response2.data), "Use HTTPS"))
}
