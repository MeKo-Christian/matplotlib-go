package main

import (
	"reflect"
	"testing"
)

func TestRenderIDsFromFlagsAcceptsSingleID(t *testing.T) {
	ids, err := renderIDsFromFlags(" basic_line ", "", false)
	if err != nil {
		t.Fatalf("renderIDsFromFlags returned error: %v", err)
	}
	if !reflect.DeepEqual(ids, []string{"basic_line"}) {
		t.Fatalf("ids = %#v, want single basic_line", ids)
	}
}

func TestRenderIDsFromFlagsAcceptsCommaSeparatedIDs(t *testing.T) {
	ids, err := renderIDsFromFlags("", "basic_line, scatter_basic ,, hist_basic ", false)
	if err != nil {
		t.Fatalf("renderIDsFromFlags returned error: %v", err)
	}
	want := []string{"basic_line", "scatter_basic", "hist_basic"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("ids = %#v, want %#v", ids, want)
	}
}

func TestRenderIDsFromFlagsRejectsMissingIDs(t *testing.T) {
	if _, err := renderIDsFromFlags("", " , ", false); err == nil {
		t.Fatal("expected missing ids to fail")
	}
}

func TestRenderIDsFromFlagsRejectsConflictingModes(t *testing.T) {
	if _, err := renderIDsFromFlags("basic_line", "scatter_basic", false); err == nil {
		t.Fatal("expected --id and --ids together to fail")
	}
	if _, err := renderIDsFromFlags("basic_line", "", true); err == nil {
		t.Fatal("expected --all and --id together to fail")
	}
}
