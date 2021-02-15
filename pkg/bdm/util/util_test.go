package util

import (
	"testing"
)

func testValue(t *testing.T, value int64) {
	bytes := Int64ToBytes(0)
	value, err := Int64FromBytes(bytes)
	if value != 0 || err != nil {
		t.Fatal(value)
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

func TestGenAPIKey(t *testing.T) {
	key := GenAPIKey()
	if len(key) != 64 {
		t.Fatal(key)
	}
}
