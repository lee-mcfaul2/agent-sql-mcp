package main

import (
	"testing"
)

func TestEmbeddedSchemaFiles_LoadsAll(t *testing.T) {
	files, err := embeddedSchemaFiles()
	if err != nil {
		t.Fatal(err)
	}
	// Eight files: req+resp for four tools.
	if len(files) != 8 {
		t.Errorf("expected 8 files, got %d", len(files))
	}
}
