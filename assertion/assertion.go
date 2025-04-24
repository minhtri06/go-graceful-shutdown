package assertion

import "testing"

func Equal[T comparable](t testing.TB, got, expect T) {
	t.Helper()
	if got != expect {
		t.Errorf("expect %v, got %v", expect, got)
	}
}

func NoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("didn't expect an error, but got one: %v", err)
	}
}
