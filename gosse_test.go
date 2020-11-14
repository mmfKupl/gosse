package gosse

import "testing"

func TestHello(t *testing.T) {
	want := "Hello World"
	if got := Hello(); got != want {
		t.Errorf("Hello() = %q, want %q", got, want)
	}
}
