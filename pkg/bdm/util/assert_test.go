package util

import "testing"

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

func AssertError(t *testing.T, err error) {
	if err == nil {
		t.Helper()
		t.Fatal("expected error but got no error")
	}
}

func Assert(t *testing.T, statement bool) {
	if !statement {
		t.Helper()
		t.Fatal("assertion failed")
	}
}

func AssertEqualString(t *testing.T, expected, value string) {
	if expected != value {
		t.Helper()
		message := "Expected " + value + " to be " + expected
		t.Fatal(message)
	}
}
