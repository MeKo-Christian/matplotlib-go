package main

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/webdemo"
)

func TestSelectedBackendIDDefaultsToWebdemoDefault(t *testing.T) {
	got, err := selectedBackendID("")
	if err != nil {
		t.Fatalf("selectedBackendID(empty) error = %v", err)
	}
	if got != webdemo.DefaultBackendID() {
		t.Fatalf("selectedBackendID(empty) = %q, want %q", got, webdemo.DefaultBackendID())
	}
}

func TestSelectedBackendIDRejectsUnknownBackend(t *testing.T) {
	if _, err := selectedBackendID("not-a-backend"); err == nil {
		t.Fatal("selectedBackendID(unknown) returned nil error")
	}
}
