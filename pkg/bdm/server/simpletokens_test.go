package server

import (
	"testing"

	"github.com/cry-inc/bdm/pkg/bdm/util"
)

func TestSimpleTokens(t *testing.T) {
	tokens := CreateSimpleTokens("read", "write", "admin")

	// Check empty tokens
	util.Assert(t, !tokens.CanRead(""))
	util.Assert(t, !tokens.CanWrite(""))
	util.Assert(t, !tokens.IsAdmin(""))

	// Check wrong tokens
	util.Assert(t, !tokens.CanRead("wrong"))
	util.Assert(t, !tokens.CanWrite("wrong"))
	util.Assert(t, !tokens.IsAdmin("wrong"))

	// Check correct tokens
	util.Assert(t, tokens.CanRead("read"))
	util.Assert(t, tokens.CanWrite("write"))
	util.Assert(t, tokens.IsAdmin("admin"))

	// Check "inherited" permissions
	util.Assert(t, tokens.CanRead("write"))
	util.Assert(t, tokens.CanWrite("admin"))
	util.Assert(t, tokens.CanRead("admin"))

	// Test free for all reading without admin permissions
	tokens = CreateSimpleTokens("", "write", "")
	util.Assert(t, tokens.CanRead(""))
	util.Assert(t, tokens.CanWrite("write"))
	util.Assert(t, tokens.CanRead("write"))
	util.Assert(t, !tokens.CanRead("wrong"))
	util.Assert(t, !tokens.IsAdmin(""))

	// User mode is not supported
	util.Assert(t, tokens.NoUserMode())
}
