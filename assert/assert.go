package assert

import (
	"errors"
	"testing"
)

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

func Error(t testing.TB, got, expect error) {
	t.Helper()
	if !errors.Is(got, expect) {
		t.Errorf("expect: %v\ngot: %v", expect, got)
	}
}
