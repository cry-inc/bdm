package util

import (
	"testing"
)

func testValue(t *testing.T, input int64) {
	bytes := Int64ToBytes(input)
	output, err := Int64FromBytes(bytes)
	t.Helper()
	Assert(t, output == input && err == nil)
}

func TestInt64Helper(t *testing.T) {
	testValue(t, 0)
	testValue(t, 1)
	testValue(t, -1)
	testValue(t, 60000000000)
	testValue(t, -60000000000)

	_, err := Int64FromBytes(nil)
	AssertError(t, err)
	_, err = Int64FromBytes([]byte{1, 2, 3, 4, 5, 6, 7})
	AssertError(t, err)
}

func TestGenerateAPIToken(t *testing.T) {
	token := GenerateAPIToken()
	Assert(t, len(token) == 64)
}
