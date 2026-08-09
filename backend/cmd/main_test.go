package main

import "testing"

func TestValidRequestID(t *testing.T) {
	if !validRequestID("request-123_A") {
		t.Fatal("validRequestID() rejected a valid identifier")
	}
	for _, value := range []string{"", "contains space", "contains/slash"} {
		if validRequestID(value) {
			t.Fatalf("validRequestID() accepted %q", value)
		}
	}
}
