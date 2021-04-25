package util

import (
	"testing"
)

func testValue(t *testing.T, input int64) {
	bytes := Int64ToBytes(input)
	output, err := Int64FromBytes(bytes)
	if output != input || err != nil {
		t.Fatal(output)
	}
}

func TestInt64Helper(t *testing.T) {
	testValue(t, 0)
	testValue(t, 1)
	testValue(t, -1)
	testValue(t, 60000000000)
	testValue(t, -60000000000)

	_, err := Int64FromBytes(nil)
	if err == nil {
		t.Fatal()
	}
	_, err = Int64FromBytes([]byte{1, 2, 3, 4, 5, 6, 7})
	if err == nil {
		t.Fatal()
	}
}

func TestGenAPIToken(t *testing.T) {
	token := GenAPIToken()
	if len(token) != 64 {
		t.Fatal(token)
	}
}
