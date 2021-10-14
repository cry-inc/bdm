package util

import "testing"

// AssertNoError will fail out if err is not nil
func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

// AssertError will fail out if err is nil
func AssertError(t *testing.T, err error) {
	if err == nil {
		t.Helper()
		t.Fatal("expected error but got no error")
	}
}

// Assert will fail out if the statement is false
func Assert(t *testing.T, statement bool) {
	if !statement {
		t.Helper()
		t.Fatal("assertion failed")
	}
}

// AssertEqualString will fail if the string does match the expectation
func AssertEqualString(t *testing.T, expected, value string) {
	if expected != value {
		t.Helper()
		message := "Expected " + value + " to be " + expected
		t.Fatal(message)
	}
}
