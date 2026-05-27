package main

import (
	"testing"
)

func TestEmbeddedSchemaFiles_LoadsAll(t *testing.T) {
	files, err := embeddedSchemaFiles()
	if err != nil {
		t.Fatal(err)
	}
	// 24 files: meta+req+resp for eight tools (original five + three
	// list_all_<table> browse tools).
	if len(files) != 24 {
		t.Errorf("expected 24 files, got %d", len(files))
	}
}
