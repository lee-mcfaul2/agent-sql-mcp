package main

import (
	"testing"
)

func TestEmbeddedSchemaFiles_LoadsAll(t *testing.T) {
	files, err := embeddedSchemaFiles()
	if err != nil {
		t.Fatal(err)
	}
	// Twelve files: meta+req+resp for four tools (meta.json added by the
	// lib-agent-prompt v1.0 schema propagation).
	if len(files) != 12 {
		t.Errorf("expected 12 files, got %d", len(files))
	}
}
